package protocol

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const OpenCloseTimeout = 30 * time.Second // stream open/close timeout

// Session represents the abstraction of transport between a client and a server.
// Session can be used synomously as a client or a server.
type Session struct {
	// Session Logger
	logger *logrus.Logger

	// Indicates if a session is a client or a server
	client bool

	// Writes request monotonic increasing
	requestID uint32

	// UDP connection that defines the transport layer
	conn *net.UDPConn

	// Channel that notifies for new streams
	chStreamAccept chan *Stream

	// Channel that notifies if session has been closed
	chDie chan struct{}

	// Mutex for streams
	streamLock sync.Mutex

	// Mapping of SID/RID to streams.
	// Stores all active streams
	streams map[string]*Stream

	// Channel for outgoing writes
	chWrites chan writeRequest

	// Defines the maximum frame size for transport
	maxFrameSize int

	// Socket errors
	chSocketReadError    chan struct{}
	chSocketWriteError   chan struct{}
	socketReadError      atomic.Value
	socketWriteError     atomic.Value
	socketReadErrorOnce  sync.Once
	socketWriteErrorOnce sync.Once

	// protocol errors
	protoError     atomic.Value
	chProtoError   chan struct{}
	protoErrorOnce sync.Once
}

type writeRequest struct {
	frame  Frame
	result chan writeResult
}

type writeResult struct {
	n   int
	err error
}

// NewSession creates a new session that defines a server or client
func NewSession(conn *net.UDPConn, client bool) *Session {
	s := new(Session)
	s.conn = conn
	s.client = client
	s.logger = logrus.New()
	s.maxFrameSize = 1500
	s.streams = make(map[string]*Stream)

	s.chDie = make(chan struct{})
	s.chWrites = make(chan writeRequest)
	s.chStreamAccept = make(chan *Stream, 1500)
	s.chSocketReadError = make(chan struct{})
	s.chSocketWriteError = make(chan struct{})
	s.chProtoError = make(chan struct{})

	return s
}

func (s *Session) Start() {
	go s.recv()
	go s.send()
}

func (s *Session) Close() error {
	var once bool
	close(s.chDie)

	if once {
		s.streamLock.Lock()
		for k := range s.streams {
			s.streams[k].sessionClose()
		}
		s.streamLock.Unlock()
		return s.conn.Close()
	} else {
		return io.ErrClosedPipe
	}
}

// Accept blocks until a new stream is created
func (s *Session) Accept() (*Stream, error) {
	select {
	case stream := <-s.chStreamAccept:
		return stream, nil
	case <-s.chDie:
		return nil, io.ErrClosedPipe
	case <-s.chSocketReadError:
		return nil, s.socketReadError.Load().(error)
	case <-s.chProtoError:
		return nil, s.protoError.Load().(error)
	}
}

func (s *Session) OpenWithExisting(addr *net.UDPAddr, old *Stream) (*Stream, error) {
	if s.IsClosed() {
		return nil, io.ErrClosedPipe
	}
	stream := NewStream(s, old.sid, s.requestID, s.maxFrameSize, addr)
	return s.open(stream)
}

func (s *Session) Open(addr *net.UDPAddr) (*Stream, error) {
	if s.IsClosed() {
		return nil, io.ErrClosedPipe
	}

	sid := uuid.New()
	stream := NewStream(s, sid[:], s.requestID, s.maxFrameSize, addr)
	return s.open(stream)
}

func (s *Session) open(stream *Stream) (*Stream, error) {
	if _, err := s.writeFrame(NewFrame(SYN, stream.sid, stream.rid, 0), time.After(OpenCloseTimeout)); err != nil {
		return nil, err
	}

	s.streamLock.Lock()
	defer s.streamLock.Unlock()
	select {
	case <-s.chDie:
		return nil, io.ErrClosedPipe
	case <-s.chSocketReadError:
		return nil, s.socketReadError.Load().(error)
	case <-s.chProtoError:
		return nil, s.protoError.Load().(error)
	default:
		s.streams[fmt.Sprintf("%v%v", stream.sid, stream.rid)] = stream
		atomic.AddUint32(&s.requestID, 1)
		return stream, nil
	}
}

func (s *Session) IsClosed() bool {
	select {
	case <-s.chDie:
		return true
	default:
		return false
	}
}

func (s *Session) recv() {
	var addr *net.UDPAddr
	var err error
	var hdr header

	for {
		b := make([]byte, 1500)
		// Read header
		_, addr, err = s.conn.ReadFromUDP(b)
		if err != nil {
			s.notifyReadError(err)
			return
		}

		copy(hdr[:], b[:HeaderSize])
		sid := hdr.StreamID()
		rid := hdr.RequestID()
		seqId := hdr.SeqId()

		sidRid := fmt.Sprintf("%v%v", sid, rid)

		switch hdr.Flag() {
		case SYN:
			s.streamLock.Lock()
			// Create new stream
			if _, ok := s.streams[sidRid]; !ok {
				stream := NewStream(s, sid, rid, s.maxFrameSize, addr)
				s.streams[sidRid] = stream
				select {
				case <-s.chDie:
				case s.chStreamAccept <- stream:
				}
			}
			s.streamLock.Unlock()
		case PSH:
			if hdr.Length() <= 0 {
				continue
			}
			s.streamLock.Lock()
			if stream, ok := s.streams[sidRid]; ok {
				newbuf := make([]byte, int(hdr.Length()))
				copy(newbuf, b[HeaderSize:HeaderSize+int(hdr.Length())])
				stream.pushBytes(seqId, newbuf)
				// atomic.AddInt32(&s.bucket, -int32(written))
				stream.notifyReadEvent()
			}
			s.streamLock.Unlock()

		case DNE:
			s.streamLock.Lock()
			if stream, ok := s.streams[sidRid]; ok {
				stream.notifyACKEvent()
			}
			s.streamLock.Unlock()
		case FIN:
			s.streamLock.Lock()
			if stream, ok := s.streams[sidRid]; ok {
				stream.fin()
				// remove blocks to on going read
				stream.notifyReadEvent()
			}
			s.streamLock.Unlock()
		case NOP:
		default:
			s.notifyProtoError(ErrInvalidProtocol)
			return
		}
	}
}

// notify the session that a stream has closed
func (s *Session) streamClosed(sid []byte, rid uint32) {
	s.streamLock.Lock()
	delete(s.streams, fmt.Sprintf("%v%v", sid, rid))
	s.streamLock.Unlock()
}

func (s *Session) notifyReadError(err error) {
	s.socketReadErrorOnce.Do(func() {
		s.socketReadError.Store(err)
		close(s.chSocketReadError)
	})
}

func (s *Session) notifyWriteError(err error) {
	s.socketWriteErrorOnce.Do(func() {
		s.socketWriteError.Store(err)
		close(s.chSocketWriteError)
	})
}

func (s *Session) notifyProtoError(err error) {
	s.protoErrorOnce.Do(func() {
		s.protoError.Store(err)
		close(s.chProtoError)
	})
}

func (s *Session) send() {
	var buf []byte
	var n int
	var err error

	// 2^16 + 7 buffer size
	buf = make([]byte, (1<<16)+HeaderSize)

	for {
		select {
		case <-s.chDie:
			return
		case request := <-s.chWrites:
			header := request.frame.Header()
			copy(buf[:HeaderSize], header[:])
			copy(buf[HeaderSize:], request.frame.Data)

			if s.client {
				s.conn.Write(buf[:HeaderSize+len(request.frame.Data)])
			} else {
				// Retrieve the stream
				stream, ok := s.streams[fmt.Sprintf("%v%v", request.frame.Sid, request.frame.Rid)]
				if ok {
					s.conn.WriteToUDP(buf[:HeaderSize+len(request.frame.Data)], stream.addr)
				} else {
					// Stream does not exist
					n, err = 0, errors.New("Stream not found. Might have been closed")
				}
			}

			n -= HeaderSize
			if n < 0 {
				n = 0
			}

			result := writeResult{
				n:   n,
				err: err,
			}

			request.result <- result
			close(request.result)

			// notify connection write error
			if err != nil {
				s.notifyWriteError(err)
				return
			}
		}
	}
}

func (s *Session) writeFrame(f Frame, deadline <-chan time.Time) (n int, err error) {
	req := writeRequest{
		frame:  f,
		result: make(chan writeResult, 1),
	}

	select {
	case s.chWrites <- req:
	case <-s.chDie:
		return 0, io.ErrClosedPipe
	case <-s.chSocketWriteError:
		return 0, s.socketWriteError.Load().(error)
	case <-deadline:
		return 0, ErrTimeout
	}

	select {
	case result := <-req.result:
		return result.n, result.err
	case <-s.chDie:
		return 0, io.ErrClosedPipe
	case <-s.chSocketWriteError:
		return 0, s.socketWriteError.Load().(error)
	case <-deadline:
		return 0, ErrTimeout
	}
}

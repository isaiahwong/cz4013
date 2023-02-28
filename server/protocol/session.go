package protocol

import (
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

const OpenCloseTimeout = 30 * time.Second // stream open/close timeout

type Session struct {
	client bool

	conn     *net.UDPConn
	capacity int

	// next stream identifier used for clients
	nextStreamID    uint32
	nextStreamIDMux sync.Mutex

	chStreamAccept chan *Stream
	// if session has been closed
	chDie  chan struct{}
	logger *logrus.Logger

	// mutex streams
	streamMux sync.Mutex
	// mapping of sid to streams.
	streams  map[uint32]*Stream
	chWrites chan writeRequest

	maxFrameSize int

	requestID uint32 // write request monotonic increasing

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
	seq    uint32
	result chan writeResult
}

type writeResult struct {
	n   int
	err error
}

func NewSession(conn *net.UDPConn, client bool) *Session {
	s := new(Session)
	s.conn = conn
	s.client = client
	s.capacity = 1024
	s.logger = logrus.New()
	s.maxFrameSize = 1024
	s.streams = make(map[uint32]*Stream)

	s.chDie = make(chan struct{})
	s.chWrites = make(chan writeRequest)
	s.chStreamAccept = make(chan *Stream, 1024)
	s.chSocketReadError = make(chan struct{})
	s.chSocketWriteError = make(chan struct{})
	s.chProtoError = make(chan struct{})

	return s
}

func (s *Session) Start() {
	go s.recvLoop()
	go s.sendLoop()
}

func (s *Session) Close() error {
	var once bool
	close(s.chDie)

	if once {
		s.streamMux.Lock()
		for k := range s.streams {
			s.streams[k].sessionClose()
		}
		s.streamMux.Unlock()
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

func (s *Session) Open(addr *net.UDPAddr) (*Stream, error) {
	if s.IsClosed() {
		return nil, io.ErrClosedPipe
	}

	// generate stream id
	s.nextStreamIDMux.Lock()

	s.nextStreamID += 2
	sid := s.nextStreamID
	s.nextStreamIDMux.Unlock()

	stream := NewStream(s, sid, s.maxFrameSize, addr)

	if _, err := s.writeFrame(NewFrame(SYN, sid), time.After(OpenCloseTimeout)); err != nil {
		return nil, err
	}

	s.streamMux.Lock()
	defer s.streamMux.Unlock()
	select {
	case <-s.chDie:
		return nil, io.ErrClosedPipe
	case <-s.chSocketReadError:
		return nil, s.socketReadError.Load().(error)
	case <-s.chProtoError:
		return nil, s.protoError.Load().(error)
	default:
		s.streams[sid] = stream
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

func (s *Session) recvLoop() {
	var addr *net.UDPAddr
	var err error
	var hdr header

	for {
		b := make([]byte, 1024)
		// Read header
		_, addr, err = s.conn.ReadFromUDP(b)
		copy(hdr[:], b[:HeaderSize])

		if err != nil {
			s.notifyReadError(err)
			return
		}

		sid := hdr.StreamID()
		switch hdr.Flag() {
		case SYN:
			s.streamMux.Lock()
			// Create new stream
			if _, ok := s.streams[sid]; !ok {
				stream := NewStream(s, sid, s.maxFrameSize, addr)
				s.streams[sid] = stream
				select {
				case <-s.chDie:
				case s.chStreamAccept <- stream:
				}
			}
			s.streamMux.Unlock()
		case PSH:
			if hdr.Length() <= 0 {
				continue
			}
			newbuf := make([]byte, int(hdr.Length()))
			copy(newbuf, b[HeaderSize:HeaderSize+int(hdr.Length())])
			s.streamMux.Lock()
			if stream, ok := s.streams[sid]; ok {
				stream.pushBytes(newbuf)
				// atomic.AddInt32(&s.bucket, -int32(written))
				stream.notifyReadEvent()
			}
			s.streamMux.Unlock()

		case ACK:
			s.streamMux.Lock()
			if stream, ok := s.streams[sid]; ok {
				stream.notifyACKEvent()
			}
			s.streamMux.Unlock()
		case FIN:
			s.streamMux.Lock()
			if stream, ok := s.streams[sid]; ok {
				stream.fin()
				// remove blocks to on going read
				stream.notifyReadEvent()
			}
			s.streamMux.Unlock()
		case NOP:
		default:
			s.notifyProtoError(ErrInvalidProtocol)
			return
		}
	}
}

// notify the session that a stream has closed
func (s *Session) streamClosed(sid uint32) {
	s.streamMux.Lock()
	delete(s.streams, sid)
	s.streamMux.Unlock()
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

func (s *Session) sendLoop() {
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
				stream, ok := s.streams[request.frame.Sid]
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
		seq:    atomic.AddUint32(&s.requestID, 1),
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

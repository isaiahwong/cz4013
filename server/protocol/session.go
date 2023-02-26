package protocol

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

const OpenCloseTimeout = 30 * time.Second // stream open/close timeout

type Session struct {
	conn     *net.UDPConn
	capacity int

	nextStreamID    uint32 // next stream identifier
	nextStreamIDMux sync.Mutex

	chStreamAccept chan *Stream
	die            chan struct{} // Session has been closed
	logger         *logrus.Logger

	streamMux sync.Mutex         // locks streams
	streams   map[uint32]*Stream // mapping of sid to streams.
	writes    chan writeRequest

	maxFrameSize int

	requestID uint32 // write request monotonic increasing
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

func NewSession(conn *net.UDPConn) *Session {
	s := new(Session)
	s.conn = conn
	s.capacity = 1024
	s.logger = logrus.New()
	s.maxFrameSize = 1024
	s.die = make(chan struct{})
	s.writes = make(chan writeRequest)
	s.streams = make(map[uint32]*Stream)
	s.chStreamAccept = make(chan *Stream, 1024)

	return s
}

func (s *Session) Start() {
	go s.recvLoop()
	go s.sendLoop()
}

func (s *Session) Close() error {
	var once bool
	close(s.die)

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
	case <-s.die:
		return nil, io.ErrClosedPipe
	}
}

func (s *Session) Open(addr *net.UDPAddr) (*Stream, error) {
	if s.IsClosed() {
		return nil, io.ErrClosedPipe
	}

	// generate stream id
	s.nextStreamIDMux.Lock()
	// if s.goAway > 0 {
	// 	s.nextStreamIDLock.Unlock()
	// 	return nil, ErrGoAway
	// }

	s.nextStreamID += 2
	sid := s.nextStreamID
	// if sid == sid%2 { // stream-id overflows
	// 	s.goAway = 1
	// 	s.nextStreamIDLock.Unlock()
	// 	return nil, ErrGoAway
	// }
	s.nextStreamIDMux.Unlock()

	stream := NewStream(s, sid, s.maxFrameSize, addr)

	if _, err := s.writeFrame(NewFrame(SYN, sid), time.After(OpenCloseTimeout)); err != nil {
		return nil, err
	}

	s.streamMux.Lock()
	defer s.streamMux.Unlock()
	select {
	case <-s.die:
		return nil, io.ErrClosedPipe
	default:
		s.streams[sid] = stream
		return stream, nil
	}
}

func (s *Session) IsClosed() bool {
	select {
	case <-s.die:
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
		copy(hdr[:], b[:headerSize])

		if err != nil {
			s.logger.WithError(err).Log(logrus.ErrorLevel, "recvLoop: Unable to read from UDP connection. Closing")
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
				case s.chStreamAccept <- stream:
				case <-s.die:
				}
			}
			s.streamMux.Unlock()
		case PSH:
			if hdr.Length() <= 0 {
				continue
			}

			newbuf := make([]byte, int(hdr.Length()))
			copy(newbuf, b[headerSize:headerSize+int(hdr.Length())])
			s.streamMux.Lock()
			if stream, ok := s.streams[sid]; ok {
				stream.pushBytes(newbuf)
				// atomic.AddInt32(&s.bucket, -int32(written))
				stream.notifyReadEvent()
			}
			s.streamMux.Unlock()
			// if _, err := io.ReadFull(s.conn, newbuf); err == nil {

			// 	fmt.Println("here")
			// 	s.streamMux.Lock()
			// 	if stream, ok := s.streams[sid]; ok {

			// 		stream.pushBytes(newbuf)
			// 		// atomic.AddInt32(&s.bucket, -int32(written))
			// 		stream.notifyReadEvent()
			// 	}
			// 	s.streamMux.Unlock()
			// } else {
			// 	fmt.Println(err)
			// 	// s.notifyReadError(err)
			// 	return
			// }

		case FIN:
			s.streamMux.Lock()
			if stream, ok := s.streams[sid]; ok {
				stream.fin()
				stream.notifyReadEvent()
			}
			s.streamMux.Unlock()
		case NOP:

		}
	}
}

// notify the session that a stream has closed
func (s *Session) streamClosed(sid uint32) {
	s.streamMux.Lock()
	delete(s.streams, sid)
	s.streamMux.Unlock()
}

func (s *Session) sendLoop() {
	var buf []byte
	var n int
	var err error

	// 2^16 + 7 buffer size
	buf = make([]byte, (1<<16)+headerSize)

	for {

		select {
		case <-s.die:
			return
		case request := <-s.writes:
			buf[0] = request.frame.Flag
			binary.LittleEndian.PutUint16(buf[1:], uint16(len(request.frame.Data)))
			binary.LittleEndian.PutUint32(buf[3:], request.frame.Sid)

			copy(buf[headerSize:], request.frame.Data)

			n, err = s.conn.Write(buf[:headerSize+len(request.frame.Data)])

			n -= headerSize
			if n < 0 {
				n = 0
			}

			result := writeResult{
				n:   n,
				err: err,
			}

			request.result <- result
			close(request.result)

			// store conn error
			// if err != nil {
			// 	s.notifyWriteError(err)
			// 	return
			// }
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
	case s.writes <- req:
	case <-s.die:
		return 0, io.ErrClosedPipe
	case <-deadline:
		return 0, ErrTimeout
	}
	select {
	case result := <-req.result:
		return result.n, result.err
	case <-s.die:
		return 0, io.ErrClosedPipe
	case <-deadline:
		return 0, ErrTimeout
	}
}

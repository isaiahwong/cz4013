package protocol

import (
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Stream struct {
	sid []byte

	session *Session

	addr *net.UDPAddr

	frameSize int

	buffers [][]byte

	bufferMux sync.Mutex

	rDeadline atomic.Value
	wDeadline atomic.Value

	// Events
	chRead chan struct{}
	chAck  chan struct{}
	chFin  chan struct{}
	chDie  chan struct{}
}

var (
	ErrInvalidProtocol = errors.New("invalid protocol")
	ErrConsumed        = errors.New("peer consumed more than sent")
	ErrGoAway          = errors.New("stream id overflows, should start a new connection")
	ErrTimeout         = errors.New("timeout")
	ErrMayBlock        = errors.New("op may block on IO")
)

func NewStream(sess *Session, sid []byte, frameSize int, addr *net.UDPAddr) *Stream {
	s := new(Stream)
	s.sid = sid
	s.frameSize = frameSize - HeaderSize
	s.session = sess
	s.addr = addr
	s.chRead = make(chan struct{})
	s.chAck = make(chan struct{}, 1) // limit to 1
	s.chFin = make(chan struct{})
	s.chDie = make(chan struct{})
	return s
}

func (s *Stream) SetReadDeadline(t time.Time) error {
	s.rDeadline.Store(t)
	s.notifyReadEvent()
	return nil
}

// SetWriteDeadline sets the write deadline as defined by
// net.Conn.SetWriteDeadline.
// A zero time value disables the deadline.
func (s *Stream) SetWriteDeadline(t time.Time) error {
	s.wDeadline.Store(t)
	return nil
}

func (s *Stream) Close() error {
	close(s.chDie)

	_, err := s.session.writeFrame(NewFrame(FIN, s.sid), time.After(OpenCloseTimeout))
	s.session.streamClosed(s.sid)
	if err != nil {
		return err
	}
	return nil
}

func (s *Stream) IsClosed() bool {
	select {
	case <-s.chDie:
		return true
	case <-s.chFin:
		return true
	default:
		return false
	}
}

// Implements io.Reader
func (s *Stream) Read(b []byte) (int, error) {
	n := 0
	for {
		n += s.read(b, n)
		err := s.waitRead()
		if err == io.EOF {
			return n, nil
		}
		if err != nil {
			return n, err
		}
	}
}

// SID returns a string representation of byte[] sid
func (s *Stream) SID() string {
	return string(s.sid)
}

func (s *Stream) read(b []byte, offset int) (n int) {
	if len(b) == 0 {
		return 0
	}

	s.bufferMux.Lock()
	if len(s.buffers) > 0 {
		n = copy(b[offset:offset+len(s.buffers[0])], s.buffers[0])
		s.buffers[0] = s.buffers[0][n:]
		// Read finish
		if len(s.buffers[0]) == 0 {
			s.buffers[0] = nil
			s.buffers = s.buffers[1:]
		}
	}
	s.bufferMux.Unlock()

	return n
}

func (s *Stream) waitRead() error {
	var timer *time.Timer
	var deadline <-chan time.Time
	if d, ok := s.rDeadline.Load().(time.Time); ok && !d.IsZero() {
		timer = time.NewTimer(time.Until(d))
		defer timer.Stop()
		deadline = timer.C
	}

	select {
	case <-s.chRead:
		return nil
	case <-s.chAck:
		if len(s.buffers) > 0 {
			s.notifyACKEvent() // if ack is consumed, notify again until buffer is empty
			return nil
		}
		return io.EOF
	case <-s.chFin:
		s.bufferMux.Lock()
		defer s.bufferMux.Unlock()
		if len(s.buffers) > 0 {
			return nil
		}
		return io.EOF
	case <-deadline:
		return ErrTimeout
	case <-s.chDie:
		return io.ErrClosedPipe
	case <-s.session.chSocketReadError:
		return s.session.socketReadError.Load().(error)
	case <-s.session.chProtoError:
		return s.session.protoError.Load().(error)
	}
}

// pushBytes append buf to buffers
func (s *Stream) pushBytes(buf []byte) (written int, err error) {
	s.bufferMux.Lock()
	defer s.bufferMux.Unlock()
	s.buffers = append(s.buffers, buf)
	return
}

// notify read event
func (s *Stream) notifyReadEvent() {
	select {
	case s.chRead <- struct{}{}:
	default:
	}
}

func (s *Stream) notifyACKEvent() {
	select {
	case s.chAck <- struct{}{}:
	default:
	}
}

func (s *Stream) Write(b []byte) (n int, err error) {
	var deadline <-chan time.Time
	if d, ok := s.wDeadline.Load().(time.Time); ok && !d.IsZero() {
		timer := time.NewTimer(time.Until(d))
		defer timer.Stop()
		deadline = timer.C
	}

	// check if stream has closed
	select {
	case <-s.chDie:
		return 0, io.ErrClosedPipe
	default:
	}

	// frame split and transmit
	sent := 0
	frame := NewFrame(PSH, s.sid)
	bts := b
	for len(bts) > 0 {
		size := len(bts)
		// Truncate to frame size
		if size > s.frameSize {
			size = s.frameSize
		}
		frame.Data = bts[:size]
		bts = bts[size:]
		n, err := s.session.writeFrame(frame, deadline)
		// s.numWritten++
		sent += n
		if err != nil {
			return sent, err
		}
	}
	n, err = s.session.writeFrame(NewFrame(ACK, s.sid), time.After(OpenCloseTimeout))
	sent += n
	// Finish write with ACK
	if err != nil {
		return sent, err
	}

	return sent, nil
}

func (s *Stream) fin() {
	close(s.chFin)
}

func (s *Stream) sessionClose() { close(s.chDie) }

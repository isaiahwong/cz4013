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
	sid uint32

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

func NewStream(sess *Session, sid uint32, frameSize int, addr *net.UDPAddr) *Stream {
	s := new(Stream)
	s.sid = sid
	s.frameSize = frameSize
	s.session = sess
	s.addr = addr
	s.chRead = make(chan struct{}, 1)
	s.chAck = make(chan struct{}, 1)
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

// Implements io.Reader
func (s *Stream) Read(b []byte) (int, error) {
	var n int
	for {
		n += s.read(b)
		flag, err := s.waitRead()

		if err != nil {
			return n, err
		}
		if flag == ACK {
			return n, nil
		}
	}
}

func (s *Stream) read(b []byte) (n int) {
	if len(b) == 0 {
		return 0
	}

	s.bufferMux.Lock()
	if len(s.buffers) > 0 {
		n = copy(b, s.buffers[0])
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

func (s *Stream) waitRead() (byte, error) {
	var timer *time.Timer
	var deadline <-chan time.Time
	if d, ok := s.rDeadline.Load().(time.Time); ok && !d.IsZero() {
		timer = time.NewTimer(time.Until(d))
		defer timer.Stop()
		deadline = timer.C
	}

	select {
	case <-s.chRead:
		return PSH, nil
	case <-s.chAck:
		return ACK, nil
	case <-s.chFin:
		s.bufferMux.Lock()
		defer s.bufferMux.Unlock()
		if len(s.buffers) > 0 {
			return FIN, nil
		}
		return NOP, io.EOF
	case <-deadline:
		return NOP, ErrTimeout
	case <-s.chDie:
		return NOP, io.ErrClosedPipe
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

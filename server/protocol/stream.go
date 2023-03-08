package protocol

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type ByteSeq struct {
	SeqId uint16
	Bytes []byte
}

type BySeq []*ByteSeq

func (a BySeq) Len() int           { return len(a) }
func (a BySeq) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySeq) Less(i, j int) bool { return a[i].SeqId < a[j].SeqId }

type Stream struct {
	rid uint32

	sid []byte

	session *Session

	addr *net.UDPAddr

	frameSize int

	buffers []*ByteSeq

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

func NewStream(sess *Session, sid []byte, rid uint32, frameSize int, addr *net.UDPAddr) *Stream {
	s := new(Stream)
	s.sid = sid
	s.rid = rid
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

// SID returns a string representation of byte[] sid
func (s *Stream) SID() []byte {
	return s.sid
}

func (s *Stream) RID() uint32 {
	return s.rid
}

// SIDRID returns the concatenation of sid rid
// Used to identify a unique stream
func (s *Stream) SIDRID() string {
	return fmt.Sprintf("%v%v", s.sid, s.rid)
}

func (s *Stream) Close() error {
	close(s.chDie)

	_, err := s.session.writeFrame(NewFrame(FIN, s.sid, s.rid, 0), time.After(OpenCloseTimeout))
	s.session.streamClosed(s.sid, s.rid)
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
	byteSeqSlice := []*ByteSeq{}

	readLoop := func() error {
		for {
			byteSeq := s.read()
			if byteSeq != nil {
				byteSeqSlice = append(byteSeqSlice, byteSeq)
			}

			err := s.waitRead()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
		}
	}

	err := readLoop()
	sort.Sort(BySeq(byteSeqSlice))
	n := 0
	for _, byteSeq := range byteSeqSlice {
		copy(b[n:n+len(byteSeq.Bytes)], byteSeq.Bytes)
		n += len(byteSeq.Bytes)
	}
	return n, err
}

func (s *Stream) read() *ByteSeq {
	s.bufferMux.Lock()
	defer s.bufferMux.Unlock()

	if len(s.buffers) == 0 {
		return nil
	}

	b := s.buffers[0]
	s.buffers = s.buffers[1:]
	return b
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
func (s *Stream) pushBytes(seqId uint16, buf []byte) (written int, err error) {
	s.bufferMux.Lock()
	defer s.bufferMux.Unlock()
	s.buffers = append(s.buffers, &ByteSeq{SeqId: seqId, Bytes: buf})
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
	seq := uint16(0)

	frame := NewFrame(PSH, s.sid, s.rid, seq)
	bts := b
	for len(bts) > 0 {
		size := len(bts)
		// Truncate to frame size
		if size > s.frameSize {
			size = s.frameSize
		}
		frame.Data = bts[:size]
		frame.SeqId = seq
		bts = bts[size:]
		n, err := s.session.writeFrame(frame, deadline)

		seq += 1
		sent += n

		if err != nil {
			return sent, err
		}
	}
	n, err = s.session.writeFrame(NewFrame(ACK, s.sid, s.rid, 0), time.After(OpenCloseTimeout))
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

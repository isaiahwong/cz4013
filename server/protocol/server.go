package protocol

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/isaiahwong/cz4013/common"
	"github.com/isaiahwong/cz4013/rpc"
	"github.com/sirupsen/logrus"
)

type Semantics byte

const (
	Unknown Semantics = iota
	AtMostOnce
	AtLeastOnce
)

func IntToSemantics(s int) Semantics {
	switch s {
	case 0:
		return AtMostOnce
	case 1:
		return AtLeastOnce
	default:
		return Unknown
	}
}

func (s Semantics) String() string {
	switch s {
	case AtMostOnce:
		return "AtMostOnce"
	case AtLeastOnce:
		return "AtLeastOnce"
	default:
		return "Unknown"
	}
}

type Server struct {
	logger *logrus.Logger
	opts   options
	rpc    *rpc.RPC
	addr   *net.UDPAddr

	dbLock     sync.Mutex
	flightRepo *rpc.FlightRepo

	historyLock sync.Mutex
	history     map[string][]byte

	lossRate int
	rand     *rand.Rand
}

// Serve starts the server with blocking call
func (s *Server) Serve() (err error) {
	// Resolve UDP address
	addr, err := net.ResolveUDPAddr("udp", s.opts.port)
	if err != nil {
		s.logger.WithError(err).Fatal("Unable to resolve UDP address")
		return
	}
	s.addr = addr

	// Creates non blocking UDP Connection
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		s.logger.WithError(err).Fatal("Unable to listen on UDP")
		return err
	}
	// Create new session
	sess := NewSession(conn, false)

	// Blocking
	s.handleSession(sess)
	return nil
}

func (s *Server) handleSession(sess *Session) {
	// Start session receive and send loop goroutines
	sess.Start()
	s.logger.Info(fmt.Sprintf("Started server on %v", s.addr))
	s.logger.Info(fmt.Sprintf("Server semantic: %v", s.opts.semantic.String()))
	s.logger.Info(fmt.Sprintf("Server loss rate: %v", s.opts.lossRate))

	defer sess.Close()
	for {
		stream, err := sess.Accept()
		if err != nil {
			s.logger.WithError(err).Error("Unable to accept stream, closing server")
			return
		}
		s.logger.Info(fmt.Sprintf("Accepted stream from %v", stream.addr))
		go s.handleRequest(stream)
	}
}

// writable - a write middleware
func (s *Server) writable(stream *Stream) func([]byte, bool) (int, error) {
	writable := func(data []byte, lossy bool) (int, error) {
		// Randomly drop packets
		if lossy && s.rand.Intn(100) < s.lossRate {
			s.logger.Info(fmt.Sprintf("Dropped packet from %v", stream.addr))
			time.Sleep(5 * time.Second)
			return 0, nil
		}
		return stream.Write(data)
	}

	atMostOnce := func(data []byte, lossy bool) (int, error) {
		s.historyLock.Lock()
		// Cache results
		sid := stream.SID()
		s.history[string(sid)] = data
		s.historyLock.Unlock()

		n, err := writable(data, lossy)
		if err != nil {
			return n, err
		}
		return n, nil
	}

	switch s.opts.semantic {
	case AtMostOnce:
		return atMostOnce
	default:
		return writable
	}
}

// readable - a read middleware
func (s *Server) readable(stream *Stream) func(time.Duration) ([]byte, error) {
	readable := func(deadline time.Duration) ([]byte, error) {
		stream.SetReadDeadline(time.Now().Add(deadline))
		// MTU TODO: parametrise
		buf := make([]byte, 65507)
		// Process requests
		n, err := stream.Read(buf)
		if err != nil {
			return nil, err
		}
		return buf[:n], nil
	}

	return readable
}

func (s *Server) handleRequest(stream *Stream) {
	defer stream.Close()
	var err error

	handleRequest := s.rpc.HandleRequest

	// atMostOnce callback
	atMostOnce := func() error {
		s.historyLock.Lock()
		sid := stream.SID()
		// Check cache
		cached, ok := s.history[string(sid)]
		s.historyLock.Unlock()

		if !ok {
			return handleRequest(
				stream.addr.IP.String(),
				s.readable(stream),
				s.writable(stream),
			)
		}

		// Return cache
		s.logger.Info(fmt.Sprintf("Returning cached result for %v", stream.addr))
		_, aErr := s.writable(stream)(cached, false)
		return aErr
	}

	// Choose semantics
	switch s.opts.semantic {
	case AtMostOnce:
		err = atMostOnce()
	default:
		err = handleRequest(
			stream.addr.IP.String(),
			s.readable(stream),
			s.writable(stream),
		)
	}

	if err != nil {
		s.logger.WithError(err).Error("Error handling request")
		return
	}
}

func newAtMostOnce(opts options) *Server {
	s := new(Server)
	s.opts = opts
	s.logger = opts.logger
	s.rpc = rpc.New(opts.flightRepo, opts.reservationRepo, opts.deadline)
	s.history = make(map[string][]byte)
	s.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	s.lossRate = opts.lossRate
	return s
}

func newAtLeastOnce(opts options) *Server {
	s := new(Server)
	s.opts = opts
	s.logger = opts.logger
	s.rpc = rpc.New(opts.flightRepo, opts.reservationRepo, opts.deadline)
	s.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	s.lossRate = opts.lossRate
	return s
}

func New(opt ...Option) *Server {
	// Default options
	opts := options{
		port:     ":8080",
		logger:   common.NewLogger(),
		semantic: AtLeastOnce,
	}
	// Apply options
	for _, o := range opt {
		o(&opts)
	}

	switch opts.semantic {
	case AtMostOnce:
		return newAtMostOnce(opts)
	default:
		return newAtLeastOnce(opts)
	}
}

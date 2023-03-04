package protocol

import (
	"fmt"
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

type Server struct {
	logger *logrus.Logger
	opts   options
	rpc    *rpc.RPC
	addr   *net.UDPAddr

	dbLock     sync.Mutex
	flightRepo *rpc.FlightRepo

	historyLock sync.Mutex
	history     map[string][]byte
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
func (s *Server) writable(stream *Stream) func([]byte) (int, error) {
	writable := func(data []byte) (int, error) {
		return stream.Write(data)
	}

	atMostOnce := func(data []byte) (int, error) {
		s.historyLock.Lock()
		defer s.historyLock.Unlock()

		n, err := writable(data)
		if err != nil {
			return n, err
		}

		// Cache results
		sid := stream.SID()
		s.history[sid] = data
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
		buf, ok := s.history[sid]
		s.historyLock.Unlock()

		if !ok {
			return handleRequest(
				stream.addr.IP.String(),
				s.readable(stream),
				s.writable(stream),
			)
		}

		// Return cache
		_, aErr := s.writable(stream)(buf)
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
	return s
}

func newAtLeastOnce(opts options) *Server {
	s := new(Server)
	s.opts = opts
	s.logger = opts.logger
	s.rpc = rpc.New(opts.flightRepo, opts.reservationRepo, opts.deadline)
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

	s := new(Server)
	s.opts = opts
	s.logger = opts.logger
	s.rpc = rpc.New(opts.flightRepo, opts.reservationRepo, opts.deadline)

	switch opts.semantic {
	case AtMostOnce:
		return newAtMostOnce(opts)
	default:
		return newAtLeastOnce(opts)
	}
}

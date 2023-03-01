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

	dbMux      sync.Mutex
	flightRepo *rpc.FlightRepo
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

func (s *Server) handleRequest(stream *Stream) {
	defer stream.Close()

	writable := func(data []byte) (int, error) {
		return stream.Write(data)
	}

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

	err := s.rpc.HandleRequest(readable, writable)
	if err != nil {
		s.logger.WithError(err).Error("Error handling request")
		return
	}
}

func New(opt ...Option) *Server {
	s := new(Server)
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

	s.opts = opts
	s.logger = opts.logger
	s.rpc = rpc.New(opts.flightRepo, opts.deadline)

	return s
}

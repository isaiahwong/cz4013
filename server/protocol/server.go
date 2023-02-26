package protocol

import (
	"net"

	"github.com/isaiahwong/cz4013/common"
	"github.com/isaiahwong/cz4013/encoding"
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
}

// Serve starts the server with blocking call
func (s *Server) Serve() (err error) {
	// Resolve UDP address
	addr, err := net.ResolveUDPAddr("udp", s.opts.port)
	if err != nil {
		s.logger.WithError(err).Fatal("Unable to resolve UDP address")
		return
	}

	// Creates non blocking UDP Connection
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		s.logger.WithError(err).Fatal("Unable to listen on UDP")
		return err
	}
	// Create new session
	sess := NewSession(conn)

	// Blocking
	s.handleSession(sess)
	return nil
}

func (s *Server) handleSession(sess *Session) {
	// Start session receive and send loop goroutines
	sess.Start()
	defer sess.Close()
	for {
		stream, err := sess.Accept()
		if err != nil {
			s.logger.WithError(err).Error("Unable to accept stream, closing server")
			return
		}
		go s.handleRequest(stream)
	}
}

func (s *Server) handleRequest(stream *Stream) {
	defer stream.Close()

	buf := make([]byte, 1024)
	// Process requests
	stream.Read(buf)
	// Unmarhsal message
	m := new(rpc.Message)
	err := encoding.Unmarshal(buf, m)
	if err != nil {
		s.logger.WithError(err).Error("Unable to unmarshal message")
		return
	}

	_ = s.rpc.HandleRequest(m)

	// Write return message
	stream.Write(buf)

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
	s.rpc = rpc.New()

	return s
}

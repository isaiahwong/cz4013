package protocol

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"github.com/isaiahwong/cz4013/common"
	"github.com/sirupsen/logrus"
)

type Server struct {
	conn   *net.UDPConn
	addr   *net.UDPAddr
	logger *logrus.Logger
}

// Serve starts the server.
// Blocking connection
func (s *Server) Serve() {
	defer s.conn.Close()
	for {

		buf, addr, err := receiveChunks(s.conn)
		if err != nil {
			log.Println("Error reading from connection:", err)
			continue
		}
		log.Printf("Received %d bytes from %s: %s", len(buf), addr, string(buf))

		// Process the request data
		response := []byte("Hello, client!")

		// Send the response back to the client
		_, err = s.conn.WriteToUDP(response, addr)
		if err != nil {
			fmt.Println("Error sending response:", err)
			return
		}
	}
}

func receiveChunks(conn *net.UDPConn) ([]byte, *net.UDPAddr, error) {
	var buf bytes.Buffer
	var addr *net.UDPAddr
	var err error
	var n int
	chunk := make([]byte, 1024)

	for {
		n, addr, err = conn.ReadFromUDP(chunk)
		if err != nil {
			return nil, addr, err
		}

		buf.Write(chunk[:n])

		if n < len(chunk) {
			break
		}
	}

	return buf.Bytes(), addr, nil
}

func New(opt ...Option) (*Server, error) {
	// Default options
	opts := options{
		port:   ":5000",
		logger: common.NewLogger(),
	}
	// Apply options
	for _, o := range opt {
		o(&opts)
	}

	// Resolve UDP address
	addr, err := net.ResolveUDPAddr("udp", opts.port)
	if err != nil {
		opts.logger.WithError(err).Fatal("Unable to resolve UDP address")
		return nil, err
	}

	// Creates non blocking UDP Connection
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		opts.logger.WithError(err).Fatal("Unable to listen on UDP")
		return nil, err
	}

	return &Server{
		conn:   conn,
		logger: opts.logger,
		addr:   addr,
	}, nil
}

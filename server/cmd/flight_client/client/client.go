package client

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/isaiahwong/cz4013/common"
	"github.com/isaiahwong/cz4013/encoding"
	"github.com/isaiahwong/cz4013/protocol"
	"github.com/isaiahwong/cz4013/rpc"
	"github.com/sirupsen/logrus"
)

type Client struct {
	opts         options
	conn         *net.UDPConn
	remoteAddr   *net.UDPAddr
	session      *protocol.Session
	logger       *logrus.Logger
	mtu          int
	retries      int
	Reservations map[string]*rpc.ReserveFlight
}

func (c *Client) open() (*protocol.Stream, error) {
	return c.session.Open(c.remoteAddr)
}

func (c *Client) openWithSid(sid []byte) (*protocol.Stream, error) {
	return c.session.OpenWithSid(c.remoteAddr, sid)
}

// sendOnly -- sends a request only
func (c *Client) sendOnly(stream *protocol.Stream, method string, query map[string]string, deadline *time.Duration) error {
	// Request
	req := &rpc.Message{
		RPC:   method,
		Query: query,
		Body:  []byte{},
	}
	b, err := encoding.Marshal(req)
	if err != nil {
		return err
	}
	stream.Write(b)
	return nil
}

// send -- sends a request and waits for a response
func (c *Client) send(stream *protocol.Stream, method string, query map[string]string, deadline *time.Duration) (*rpc.Message, *protocol.Stream, error) {
	var err error
	var m *rpc.Message
	var n int
	tries := 0

	for tries <= c.retries {
		tries++
		// Log if it's a retry
		if tries > 1 {
			stream.Close()
			c.logger.WithFields(logrus.Fields{
				"method": method,
				"query":  query,
				"tries":  tries,
			}).Info("Retrying in 1s")
			time.Sleep(1 * time.Second)
			stream, err = c.openWithSid([]byte(stream.SID()))

		}

		// Request
		err = c.sendOnly(stream, method, query, deadline)
		if err != nil {
			continue
		}

		// Response
		res := make([]byte, c.mtu)
		if deadline != nil {
			stream.SetReadDeadline(time.Now().Add(*deadline))
		}
		n, err = stream.Read(res)
		if err != nil && err != io.EOF {
			continue
		}

		m = new(rpc.Message)
		if err = encoding.Unmarshal(res[:n], m); err != nil && err != io.EOF {
			continue
		}
		break
	}
	if err != nil {
		return nil, stream, err
	}
	if m == nil {
		return nil, stream, errors.New("No response received from server.")
	}

	return m, stream, nil
}

func (c *Client) Start() (err error) {
	// Create a UDP address for the server
	c.remoteAddr, err = net.ResolveUDPAddr("udp", c.opts.addr)
	if err != nil {
		return
	}

	// Create a UDP connection to the server
	c.conn, err = net.DialUDP("udp", nil, c.remoteAddr)
	if err != nil {
		return
	}

	c.session = protocol.NewSession(c.conn, true)
	c.session.Start()
	return
}

func New(opt ...Option) *Client {
	// Default options
	opts := options{
		addr:     "localhost:8080",
		logger:   common.NewLogger(),
		deadline: time.Second * 5,
		retries:  0,
	}

	// Apply options
	for _, o := range opt {
		o(&opts)
	}

	return &Client{
		opts:         opts,
		logger:       opts.logger,
		mtu:          65507,
		retries:      opts.retries,
		Reservations: make(map[string]*rpc.ReserveFlight),
	}
}

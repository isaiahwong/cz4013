package client

import (
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
	Reservations map[string]*rpc.ReserveFlight
}

func (c *Client) open() (*protocol.Stream, error) {
	return c.session.Open(c.remoteAddr)
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
func (c *Client) send(stream *protocol.Stream, method string, query map[string]string, deadline *time.Duration) (*rpc.Message, error) {
	// Request
	err := c.sendOnly(stream, method, query, deadline)
	if err != nil {
		return nil, err
	}
	// Response
	res := make([]byte, c.mtu)
	if deadline != nil {
		stream.SetReadDeadline(time.Now().Add(*deadline))
	}
	n, err := stream.Read(res)
	if err != nil && err != io.EOF {
		return nil, err
	}
	m := new(rpc.Message)
	if err = encoding.Unmarshal(res[:n], m); err != nil && err != io.EOF {
		return nil, err
	}
	return m, err
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
	}

	// Apply options
	for _, o := range opt {
		o(&opts)
	}

	return &Client{
		opts:         opts,
		logger:       opts.logger,
		mtu:          65507,
		Reservations: make(map[string]*rpc.ReserveFlight),
	}
}

package rpc

import (
	"errors"
	"time"

	"github.com/isaiahwong/cz4013/encoding"
	"github.com/sirupsen/logrus"
)

var (
	ErrNotFound = errors.New("RPC method not found")
	ErrMarshal  = errors.New("Unable to unmarshal message")
)

type RPC struct {
	logger   *logrus.Logger
	deadline time.Duration
}

// function to handle the request
type Writable func([]byte) (int, error)
type Readable func(time.Duration) ([]byte, error)

// HandleRequest handles the request from the client
func (r *RPC) HandleRequest(read Readable, write Writable) error {
	// Read message
	buf, err := read(r.deadline)
	if err != nil {
		return err
	}

	// Unmarhsal message
	m := new(Message)
	err = encoding.Unmarshal(buf, m)
	if err != nil {
		return r.errorMessage(ErrMarshal, "", read, write)
	}

	return r.router(m, read, write)
}

func (r *RPC) router(m *Message, read Readable, write Writable) error {
	if m.RPC == "FindFlights" {
		return r.FindFlights(m, read, write)
	}

	return r.errorMessage(ErrNotFound, "RPC method not found", read, write)
}

func (r *RPC) errorMessage(err error, body string, read Readable, write Writable) error {
	b, err := encoding.Marshal(NewError(err, body))
	if err != nil {
		return err
	}
	if _, err := write(b); err != nil {
		return err
	}
	return nil
}

func New(deadline time.Duration) *RPC {
	return &RPC{
		logger:   logrus.New(),
		deadline: deadline,
	}
}

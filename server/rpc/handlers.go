package rpc

import (
	"fmt"
	"time"

	"github.com/isaiahwong/cz4013/encoding"
	"github.com/sirupsen/logrus"
)

type RPC struct {
	logger   *logrus.Logger
	deadline time.Duration
}

// function to handle the request
type Writable func([]byte) (int, error)
type Readable func(time.Duration) ([]byte, error)

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
		return fmt.Errorf("unable to unmarshal message: %w", err)
	}
	fmt.Println(m)

	// _ = s.rpc.HandleRequest(m)
	return nil
}

func (r *RPC) router() {

}

func New(deadline time.Duration) *RPC {
	return &RPC{
		logger:   logrus.New(),
		deadline: deadline,
	}
}

package client

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/isaiahwong/cz4013/encoding"
	"github.com/isaiahwong/cz4013/rpc"
)

func (c *Client) ReserveFlight(id string, seats int) (*rpc.ReserveFlight, error) {
	method := "ReserveFlight"
	// Open a new stream
	stream, err := c.open()
	if err != nil {
		return nil, err
	}

	req := map[string]string{
		"id":    id,
		"seats": fmt.Sprint(seats),
	}

	res, err := c.send(stream, method, req, &c.opts.deadline)
	if res.Error != nil {
		return nil, errors.New(res.Error.Body)
	}

	reservation := new(rpc.ReserveFlight)

	if err = encoding.Unmarshal(res.Body, reservation); err != nil && err != io.EOF {
		return nil, err
	}
	return reservation, stream.Close()
}

func (c *Client) MonitorUpdates(duration time.Duration) error {
	method := "MonitorUpdates"
	// Open a new stream
	stream, err := c.open()
	defer stream.Close()

	if err != nil {
		return err
	}

	req := map[string]string{
		"timestamp": fmt.Sprintf("%v", time.Now().Add(duration).Unix()*1000),
	}

	err = c.sendOnly(stream, method, req, &c.opts.deadline)

	for !stream.IsClosed() {
		res := make([]byte, c.mtu)
		n, err := stream.Read(res)

		if err != nil && err != io.EOF {
			return nil
		}

		m := new(rpc.Message)
		if err = encoding.Unmarshal(res[:n], m); err != nil && err != io.EOF {
			panic(err)
		}

		flight := new(rpc.Flight)
		if err = encoding.Unmarshal(m.Body, flight); err != nil && err != io.EOF {
			return err
		}
		fmt.Println("New Updated flight: ", flight)
	}

	return nil
}

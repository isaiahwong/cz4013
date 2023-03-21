package client

import (
	"errors"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/isaiahwong/cz4013/encoding"
	"github.com/isaiahwong/cz4013/rpc"
)

type Placeholder struct {
	Time time.Time
}

// FindFlights is a rpc method that finds flights by source and destination
func (c *Client) FindFlights(source string, destination string) ([]*rpc.Flight, error) {
	method := "FindFlights"
	// Open a new stream
	stream, err := c.open()
	if err != nil {
		return nil, err
	}

	req := map[string]string{
		"source":      source,
		"destination": destination,
	}
	res, stream, err := c.send(stream, method, req, &c.opts.deadline)
	if err != nil {
		return nil, err
	}
	if res.Error != nil {
		return nil, errors.New(res.Error.Body)
	}

	flights := []*rpc.Flight{}
	if err = encoding.Unmarshal(res.Body, &flights); err != nil && err != io.EOF {
		return nil, err
	}
	return flights, stream.Close()
}

// FindFlight is a rpc method that finds a flight by id
func (c *Client) FindFlight(id string) (*rpc.Flight, error) {
	method := "FindFlight"
	// Open a new stream
	stream, err := c.open()
	if err != nil {
		return nil, err
	}

	req := map[string]string{
		"id": id,
	}

	res, stream, err := c.send(stream, method, req, &c.opts.deadline)
	if err != nil {
		return nil, err
	}
	if res.Error != nil {
		return nil, errors.New(res.Error.Body)
	}

	flight := new(rpc.Flight)
	if err = encoding.Unmarshal(res.Body, flight); err != nil && err != io.EOF {
		return nil, err
	}
	return flight, stream.Close()
}

// ReserveFlight is a rpc method that reserves a flight by id and number of seats
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

	res, stream, err := c.send(stream, method, req, &c.opts.deadline)
	if err != nil {
		return nil, err
	}
	if res.Error != nil {
		return nil, errors.New(res.Error.Body)
	}

	reservation := new(rpc.ReserveFlight)
	if err = encoding.Unmarshal(res.Body, reservation); err != nil && err != io.EOF {
		return nil, err
	}
	return reservation, stream.Close()
}

// CheckInFlight is a rpc method that checks in a flight by reservation id
func (c *Client) CheckInFlight(id string) (*rpc.ReserveFlight, error) {
	method := "CheckInFlight"
	// Open a new stream
	stream, err := c.open()
	if err != nil {
		return nil, err
	}

	req := map[string]string{
		"id": id,
	}

	res, stream, err := c.send(stream, method, req, &c.opts.deadline)
	if err != nil {
		return nil, err
	}
	if res.Error != nil {
		return nil, errors.New(res.Error.Body)
	}

	reservation := new(rpc.ReserveFlight)
	if err = encoding.Unmarshal(res.Body, reservation); err != nil && err != io.EOF {
		return nil, err
	}
	return reservation, stream.Close()
}

func (c *Client) GetMeals() ([]*rpc.Food, error) {
	method := "GetMeals"
	// Open a new stream
	stream, err := c.open()
	if err != nil {
		return nil, err
	}
	req := map[string]string{}
	res, stream, err := c.send(stream, method, req, &c.opts.deadline)
	if err != nil {
		return nil, err
	}
	if res.Error != nil {
		return nil, errors.New(res.Error.Body)
	}

	meals := []*rpc.Food{}
	if err = encoding.Unmarshal(res.Body, &meals); err != nil && err != io.EOF {
		return nil, err
	}
	return meals, stream.Close()
}

func (c *Client) AddMeals(id string, mealId string) (*rpc.ReserveFlight, error) {
	method := "AddMeals"
	// Open a new stream
	stream, err := c.open()
	if err != nil {
		return nil, err
	}
	req := map[string]string{
		"id":      id,
		"meal_id": mealId,
	}
	res, stream, err := c.send(stream, method, req, &c.opts.deadline)
	if err != nil {
		return nil, err
	}
	if res.Error != nil {
		return nil, errors.New(res.Error.Body)
	}

	reservation := new(rpc.ReserveFlight)
	if err = encoding.Unmarshal(res.Body, reservation); err != nil && err != io.EOF {
		return nil, err
	}
	return reservation, stream.Close()
}

// CancelFlight is a rpc method that cancels a flight by reservation id
func (c *Client) CancelFlight(id string) (*rpc.ReserveFlight, error) {
	method := "CancelFlight"
	// Open a new stream
	stream, err := c.open()
	if err != nil {
		return nil, err
	}

	req := map[string]string{
		"id": id,
	}

	res, stream, err := c.send(stream, method, req, &c.opts.deadline)
	if err != nil {
		return nil, err
	}
	if res.Error != nil {
		return nil, errors.New(res.Error.Body)
	}

	// remove reservation
	delete(c.Reservations, id)

	reserveFlight := new(rpc.ReserveFlight)
	if err = encoding.Unmarshal(res.Body, reserveFlight); err != nil && err != io.EOF {
		return nil, err
	}
	return reserveFlight, stream.Close()
}

// MonitorUpdates is a rpc method that monitors updates for a duration.
// The method is a blocking call
func (c *Client) MonitorUpdates(duration time.Duration, interruptCh chan Placeholder) error {
	method := "MonitorUpdates"
	// Open a new stream
	stream, err := c.open()
	defer stream.Close()

	dataCh := make(chan []byte, 1)

	if err != nil {
		return err
	}

	read := func() {
		for {
			res := make([]byte, c.mtu)
			n, err := stream.Read(res)

			if err != nil && err != io.EOF {
				if err == io.ErrClosedPipe {
					return
				}
				c.logger.WithError(err).Error(method)
				return
			}

			if err == io.EOF {
				return
			}

			m := new(rpc.Message)
			if err = encoding.Unmarshal(res[:n], m); err != nil && err != io.EOF {
				c.logger.WithError(err).Error(method)
				return
			}
			select {
			case dataCh <- m.Body:
			}
		}
	}

	req := map[string]string{
		"timestamp": fmt.Sprintf("%v", time.Now().Add(duration).Unix()*1000),
	}
	err = c.sendOnly(stream, method, req, &c.opts.deadline)

	// Listen on goroutine
	go read()
	for !stream.IsClosed() {
		var body []byte
		select {
		case <-interruptCh:
			return nil
		case body = <-dataCh:
		}

		if body == nil || len(body) <= 0 {
			continue
		}

		flight := new(rpc.Flight)
		if err = encoding.Unmarshal(body, flight); err != nil && err != io.EOF {
			return err
		}
		fmt.Println("New Updated flight")
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 4, ' ', tabwriter.TabIndent)
		fmt.Println("Flight")
		fmt.Fprintln(w, flight.String())
	}
	c.logger.Info("Monitor ended")
	return nil
}

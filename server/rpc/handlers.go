package rpc

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/isaiahwong/cz4013/encoding"
	"github.com/sirupsen/logrus"
)

var (
	ErrNotFound = errors.New("RPC method not found")
	ErrMarshal  = errors.New("Unable to unmarshal message")
)

type RPC struct {
	logger          *logrus.Logger
	deadline        time.Duration
	flightRepo      *FlightRepo
	reservationRepo *ReservationRepo

	reserveMux sync.Mutex

	chFlightUpdates chan *Flight
}

// function to handle the request
type Writable func([]byte) (int, error)
type Readable func(time.Duration) ([]byte, error)

// HandleRequest handles the request from the client
func (r *RPC) HandleRequest(addr string, read Readable, write Writable) error {
	// Read message
	buf, err := read(r.deadline)
	if err != nil {
		return err
	}

	// Unmarhsal message
	m := new(Message)
	err = encoding.Unmarshal(buf, m)
	if err != nil {
		return r.error("", ErrMarshal, "", read, write)
	}

	r.logger.Info(fmt.Sprintf("[%v] - %v", m.RPC, addr))
	return r.router(m, read, write)
}

func (r *RPC) router(m *Message, read Readable, write Writable) error {
	if m.RPC == "FindFlights" {
		return r.FindFlights(m, read, write)
	} else if m.RPC == "FindFlight" {
		return r.FindFlight(m, read, write)
	} else if m.RPC == "ReserveFlight" {
		return r.ReserveFlight(m, read, write)
	} else if m.RPC == "MonitorUpdates" {
		return r.MonitorUpdates(m, read, write)
	} else if m.RPC == "CancelFlight" {
		return r.CancelFlight(m, read, write)
	}

	return r.error(m.RPC, ErrNotFound, "RPC method not found", read, write)
}

func (r *RPC) error(method string, err error, body string, read Readable, write Writable) error {
	b, err := encoding.Marshal(NewError(method, err, body))
	if err != nil {
		return err
	}
	if _, err := write(b); err != nil {
		return err
	}
	return nil
}

func (r *RPC) ok(rpc string, body []byte, read Readable, write Writable) error {
	message := NewMessage(rpc, body)
	b, err := encoding.Marshal(message)
	if err != nil {
		return err
	}
	if _, err := write(b); err != nil {
		return err
	}
	return nil
}

func New(f *FlightRepo, r *ReservationRepo, deadline time.Duration) *RPC {
	if f == nil {
		panic("flightRepo cannot be nil")
	}
	if r == nil {
		panic("reservationRepo cannot be nil")
	}
	return &RPC{
		logger:          logrus.New(),
		deadline:        deadline,
		flightRepo:      f,
		reservationRepo: r,
		chFlightUpdates: make(chan *Flight),
	}
}

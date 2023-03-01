package rpc

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/isaiahwong/cz4013/common"

	"github.com/isaiahwong/cz4013/encoding"
)

var (
	ErrNoFlightsFound = errors.New("No flights found")
	ErrNoFlightFound  = errors.New("Flight not found")
	ErrInvalidParams  = errors.New("Invalid query params")
	ErrFailToReserve  = errors.New("Failed to reserve")
)

func (r *RPC) FindFlights(m *Message, read Readable, write Writable) error {
	method := "FindFlights"
	flights, err := r.flightRepo.GetAll()
	if err != nil {
		return r.errorMessage(method, err, "", read, write)
	}

	src, ok := m.Query["source"]
	if !ok || src == "" {
		return r.errorMessage(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", src), read, write)
	}
	dest, ok := m.Query["destination"]
	if !ok || dest == "" {
		return r.errorMessage(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", dest), read, write)
	}

	filteredFlights := []*Flight{}

	sourceCriteria := func(q string, f *Flight) bool {
		exp := r.flightRepo.CreateRegexp(q)
		return exp.FindAllString(f.Source, -1) != nil
	}

	destCriteria := func(q string, f *Flight) bool {
		exp := r.flightRepo.CreateRegexp(q)
		return exp.FindAllString(f.Destination, -1) != nil
	}

	for _, f := range flights {
		if sourceCriteria(src, f) && destCriteria(dest, f) {
			filteredFlights = append(filteredFlights, f)
		}
	}

	for _, f := range filteredFlights {
		fmt.Println(*f)
	}

	b, err := encoding.Marshal(filteredFlights)
	if err != nil {
		return r.errorMessage(method, err, "", read, write)
	}

	return r.ok(method, b, read, write)
}

func (r *RPC) FindFlight(m *Message, read Readable, write Writable) error {
	method := "FindFlight"
	id, ok := m.Query["id"]
	if !ok || id == "" {
		return r.errorMessage(method, ErrInvalidParams, "", read, write)
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return r.errorMessage(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", idInt), read, write)
	}

	flight, err := r.flightRepo.FindByID(int32(idInt))
	if err != nil {
		return r.errorMessage(method, err, "", read, write)
	}
	if flight == nil {
		return r.errorMessage(method, ErrNoFlightFound, fmt.Sprintf("No flights found with %v", idInt), read, write)
	}

	b, err := encoding.Marshal(flight)
	if err != nil {
		return r.errorMessage(method, err, "", read, write)
	}

	return r.ok(method, b, read, write)
}

func (r *RPC) ReserveFlight(m *Message, read Readable, write Writable) error {
	method := "ReserveFlight"
	r.reserveMux.Lock()
	r.reserveMux.Unlock()

	id, ok := m.Query["id"]
	if !ok || id == "" {
		return r.errorMessage(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", id), read, write)
	}

	seats, ok := m.Query["seats"]
	if !ok || id == "" {
		return r.errorMessage(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", id), read, write)
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return r.errorMessage(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", idInt), read, write)
	}

	seatsInt, err := strconv.ParseInt(seats, 10, 64)
	if err != nil || seatsInt <= 0 {
		return r.errorMessage(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", seatsInt), read, write)
	}
	seatsInt32 := int32(seatsInt)

	flight, err := r.flightRepo.FindByID(int32(idInt))
	if err != nil {
		return r.errorMessage(method, err, "", read, write)
	}
	if flight == nil {
		return r.errorMessage(method, ErrNoFlightFound, fmt.Sprintf("No flights found with %v", idInt), read, write)
	}

	if flight.SeatAvailablity-seatsInt32 < 0 {
		return r.errorMessage(method, ErrFailToReserve, fmt.Sprintf("Not enough seats to reserve for flight %v", idInt), read, write)
	}

	flight.SeatAvailablity -= seatsInt32
	if err = r.flightRepo.Update(flight); err != nil {
		return r.errorMessage(method, ErrFailToReserve, err.Error(), read, write)
	}

	select {
	case r.chFlightUpdates <- flight:
	default:
	}

	b, err := encoding.Marshal(flight)
	if err != nil {
		return r.errorMessage(method, err, "", read, write)
	}

	return r.ok(method, b, read, write)
}

func (r *RPC) MonitorUpdates(m *Message, read Readable, write Writable) error {
	method := "MonitorUpdates"
	t, ok := m.Query["timestamp"]
	if !ok || t == "" {
		return r.errorMessage(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", t), read, write)
	}

	monitorInterval, err := common.StrToUnixTime(t)
	if err != nil {
		return r.errorMessage(method, ErrInvalidParams, err.Error(), read, write)
	}

	duration := time.Until(*monitorInterval)
	for {
		select {
		case flight := <-r.chFlightUpdates:
			b, err := encoding.Marshal(flight)
			if err != nil {
				return r.errorMessage(method, err, "", read, write)
			}
			r.ok(method, b, read, write)
		case <-time.After(duration):
			return r.ok(method, []byte{}, read, write)

		}
	}
}

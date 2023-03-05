package rpc

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/isaiahwong/cz4013/common"

	"github.com/isaiahwong/cz4013/encoding"
)

var (
	ErrNoFlightsFound       = errors.New("Flights not found")
	ErrNoFlightFound        = errors.New("Flight not found")
	ErrNoReserveFlightFound = errors.New("Reservation not found")
	ErrMealsNotFound        = errors.New("Meals not found")
	ErrInvalidParams        = errors.New("Invalid query params")
	ErrFailToReserve        = errors.New("Failed to reserve")
	ErrInternalError        = errors.New("Internal Error")
)

func (r *RPC) FindFlights(m *Message, read Readable, write Writable) error {
	method := "FindFlights"
	flights, err := r.flightRepo.GetAll()
	if err != nil {
		return r.error(method, err, "", read, write)
	}

	src, ok := m.Query["source"]
	if !ok || src == "" {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", src), read, write)
	}
	dest, ok := m.Query["destination"]
	if !ok || dest == "" {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", dest), read, write)
	}

	filteredFlights := []*Flight{}

	sourceCriteria := func(q string, f *Flight) bool {
		exp := common.CreateRegexp(q)
		return exp.FindAllString(f.Source, -1) != nil
	}

	destCriteria := func(q string, f *Flight) bool {
		exp := common.CreateRegexp(q)
		return exp.FindAllString(f.Destination, -1) != nil
	}

	for _, f := range flights {
		if sourceCriteria(src, f) && destCriteria(dest, f) {
			filteredFlights = append(filteredFlights, f)
		}
	}

	b, err := encoding.Marshal(filteredFlights)
	if err != nil {
		return r.error(method, err, "", read, write)
	}

	return r.ok(method, b, read, write)
}

func (r *RPC) FindFlight(m *Message, read Readable, write Writable) error {
	method := "FindFlight"
	id, ok := m.Query["id"]
	if !ok || id == "" {
		return r.error(method, ErrInvalidParams, "", read, write)
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", idInt), read, write)
	}

	flight, err := r.flightRepo.FindByID(int32(idInt))
	if err != nil {
		return r.error(method, err, "", read, write)
	}
	if flight == nil {
		return r.error(method, ErrNoFlightFound, fmt.Sprintf("No flights found with %v", idInt), read, write)
	}

	b, err := encoding.Marshal(flight)
	if err != nil {
		return r.error(method, err, "", read, write)
	}

	return r.ok(method, b, read, write)
}

func (r *RPC) ReserveFlight(m *Message, read Readable, write Writable) error {
	method := "ReserveFlight"
	r.reserveMux.Lock()
	r.reserveMux.Unlock()

	id, ok := m.Query["id"]
	if !ok || id == "" {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", id), read, write)
	}

	seats, ok := m.Query["seats"]
	if !ok || id == "" {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", id), read, write)
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", idInt), read, write)
	}

	seatsInt, err := strconv.ParseInt(seats, 10, 64)
	if err != nil || seatsInt <= 0 {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", seatsInt), read, write)
	}
	seatsInt32 := int32(seatsInt)

	flight, err := r.flightRepo.FindByID(int32(idInt))
	if err != nil {
		return r.error(method, err, "", read, write)
	}
	if flight == nil {
		return r.error(method, ErrNoFlightFound, fmt.Sprintf("No flights found with %v", idInt), read, write)
	}

	if flight.SeatAvailablity-seatsInt32 < 0 {
		return r.error(method, ErrFailToReserve, fmt.Sprintf("Not enough seats to reserve for flight %v", idInt), read, write)
	}

	reserve := &ReserveFlight{
		ID:           uuid.NewString(),
		Flight:       flight,
		SeatReserved: seatsInt32,
		CheckIn:      false,
	}

	flight.SeatAvailablity -= seatsInt32
	if err = r.flightRepo.Update(flight); err != nil {
		return r.error(method, ErrFailToReserve, err.Error(), read, write)
	}

	if err = r.reservationRepo.Insert(reserve); err != nil {
		return r.error(method, ErrFailToReserve, err.Error(), read, write)
	}

	r.broadcastFlights(flight)

	b, err := encoding.Marshal(reserve)
	if err != nil {
		return r.error(method, ErrInternalError, err.Error(), read, write)
	}

	return r.ok(method, b, read, write)
}

func (r *RPC) CheckInFlight(m *Message, read Readable, write Writable) error {
	method := "CheckInFlight"
	r.reserveMux.Lock()
	r.reserveMux.Unlock()

	id, ok := m.Query["id"]
	if !ok || id == "" {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", id), read, write)
	}

	// Retrieve reservation
	rf, err := r.reservationRepo.FindByID(id)
	if err != nil {
		return r.error(method, ErrNoReserveFlightFound, "", read, write)
	}

	marshal := func(reserve *ReserveFlight) error {
		b, err := encoding.Marshal(rf)
		if err != nil {
			return r.error(method, ErrInternalError, err.Error(), read, write)
		}
		return r.ok(method, b, read, write)
	}

	if rf.CheckIn {
		return marshal(rf)
	}

	rf.CheckIn = true
	if err = r.reservationRepo.Update(rf); err != nil {
		return r.error(method, ErrInternalError, err.Error(), read, write)
	}
	return marshal(rf)
}

func (r *RPC) GetMeals(read Readable, write Writable) error {
	method := "GetMeals"
	meals := GetFood()

	// convert meals to list
	mealList := []*Food{}
	for _, meal := range meals {
		mealList = append(mealList, meal)
	}

	b, err := encoding.Marshal(mealList)
	if err != nil {
		return r.error(method, ErrInternalError, err.Error(), read, write)
	}

	return r.ok(method, b, read, write)
}

// AddMeals is non Idempotent
func (r *RPC) AddMeals(m *Message, read Readable, write Writable) error {
	method := "AddMeals"
	r.reserveMux.Lock()
	r.reserveMux.Unlock()

	id, ok := m.Query["id"]
	if !ok || id == "" {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", id), read, write)
	}

	mealIdStr, ok := m.Query["meal_id"]
	if !ok || id == "" {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", id), read, write)
	}

	mealId, err := strconv.ParseInt(mealIdStr, 10, 64)
	if err != nil {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", id), read, write)
	}

	// Retrieve reservation
	rf, err := r.reservationRepo.FindByID(id)
	if rf == nil {
		return r.error(method, ErrNoReserveFlightFound, fmt.Sprintf("No reservation found with %v", id), read, write)
	}

	meals := GetFood()
	meal, ok := meals[int64(mealId)]
	if !ok {
		return r.error(method, ErrMealsNotFound, fmt.Sprintf("Meal not found with %v", mealIdStr), read, write)
	}

	rf.Meals = append(rf.Meals, meal)
	if err = r.reservationRepo.Update(rf); err != nil {
		return r.error(method, ErrInternalError, err.Error(), read, write)
	}

	b, err := encoding.Marshal(rf)
	if err != nil {
		return r.error(method, ErrInternalError, err.Error(), read, write)
	}
	return r.ok(method, b, read, write)
}

// CancelFlight is a Idempotent
func (r *RPC) CancelFlight(m *Message, read Readable, write Writable) error {
	method := "CancelFlight"
	r.reserveMux.Lock()
	r.reserveMux.Unlock()

	id, ok := m.Query["id"]
	if !ok || id == "" {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", id), read, write)
	}

	// Retrieve reservation
	rf, err := r.reservationRepo.FindByID(id)
	if err != nil {
		return r.error(method, ErrInternalError, err.Error(), read, write)
	}
	if rf == nil {
		return r.error(method, ErrNoReserveFlightFound, fmt.Sprintf("No reservation found with %v", id), read, write)
	}

	if rf.CheckIn {
		return r.error(method, ErrNoReserveFlightFound, "Can't cancel. Reservation checked in", read, write)
	}

	// Retrieve flight
	flight, err := r.flightRepo.FindByID(rf.Flight.ID)
	if err != nil {
		return r.error(method, ErrInternalError, err.Error(), read, write)
	}
	if flight == nil {
		return r.error(method, ErrNoFlightFound, "No flights associated with reserve flight ", read, write)
	}

	// Delete reservation
	err = r.reservationRepo.Delete(rf)
	if err != nil {
		return r.error(method, ErrInternalError, err.Error(), read, write)
	}

	// Update seats availability
	flight.SeatAvailablity += rf.SeatReserved
	if err = r.flightRepo.Update(flight); err != nil {
		return r.error(method, ErrInternalError, err.Error(), read, write)
	}

	b, err := encoding.Marshal(flight)
	if err != nil {
		return r.error(method, ErrInternalError, err.Error(), read, write)
	}

	r.broadcastFlights(flight)

	return r.ok(method, b, read, write)
}

func (r *RPC) MonitorUpdates(m *Message, read Readable, write Writable) error {
	method := "MonitorUpdates"
	t, ok := m.Query["timestamp"]
	if !ok || t == "" {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", t), read, write)
	}

	monitorInterval, err := common.StrToUnixTime(t)
	if err != nil {
		return r.error(method, ErrInvalidParams, err.Error(), read, write)
	}

	// Create channel
	r.chFlightUpdatesMux.Lock()
	fCh := make(chan *Flight)
	r.chFlightUpdates = append(r.chFlightUpdates, fCh)
	r.chFlightUpdatesMux.Unlock()

	duration := time.Until(*monitorInterval)
	for {
		select {
		case flight := <-fCh:

			b, err := encoding.Marshal(flight)
			if err != nil {
				return r.error(method, err, "", read, write)
			}
			r.ok(method, b, read, write)
		case <-time.After(duration):
			return r.ok(method, []byte{}, read, write)

		}
	}
}

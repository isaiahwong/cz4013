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

// FindFlights finds flights from source to destination. This is an idempotent method
// Takes in `source` and `destination` as its query params
func (r *RPC) FindFlights(m *Message, read Readable, write Writable) error {
	method := "FindFlights"
	lossy := true

	// Retrieve all flights
	flights, err := r.flightRepo.GetAll()
	if err != nil {
		return r.error(method, err, "", read, write)
	}

	// Process query params
	src, ok := m.Query["source"]
	if !ok || src == "" {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", src), read, write)
	}
	dest, ok := m.Query["destination"]
	if !ok || dest == "" {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", dest), read, write)
	}

	filteredFlights := []*Flight{}

	// predicate function for source
	sourceCriteria := func(q string, f *Flight) bool {
		exp := common.CreateRegexp(q)
		return exp.FindAllString(f.Source, -1) != nil
	}

	// predicate function for destination
	destCriteria := func(q string, f *Flight) bool {
		exp := common.CreateRegexp(q)
		return exp.FindAllString(f.Destination, -1) != nil
	}

	// flight flights base on predicates
	for _, f := range flights {
		if sourceCriteria(src, f) && destCriteria(dest, f) {
			filteredFlights = append(filteredFlights, f)
		}
	}

	if len(filteredFlights) == 0 {
		return r.error(method, ErrNoFlightsFound, fmt.Sprintf("No flights found from %v to %v", src, dest), read, write)
	}

	// Marshal response
	b, err := encoding.Marshal(filteredFlights)
	if err != nil {
		return r.error(method, err, "", read, write)
	}
	return r.ok(method, b, lossy, read, write)
}

// FindFlight finds a flight by `id` in query params. This is an idempotent method
func (r *RPC) FindFlight(m *Message, read Readable, write Writable) error {
	method := "FindFlight"
	lossy := true

	id, ok := m.Query["id"]
	if !ok || id == "" {
		return r.error(method, ErrInvalidParams, "", read, write)
	}

	// Converts query param to int
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", idInt), read, write)
	}

	// Finds a flight by id
	flight, err := r.flightRepo.FindByID(int32(idInt))
	if err != nil {
		return r.error(method, err, "", read, write)
	}
	if flight == nil {
		return r.error(method, ErrNoFlightFound, fmt.Sprintf("No flights found with %v", idInt), read, write)
	}

	// Marshal response
	b, err := encoding.Marshal(flight)
	if err != nil {
		return r.error(method, err, "", read, write)
	}
	return r.ok(method, b, lossy, read, write)
}

// ReserveFlight reserves a flight by `id` in query params. This is a non-idempotent method
func (r *RPC) ReserveFlight(m *Message, read Readable, write Writable) error {
	method := "ReserveFlight"
	lossy := true
	r.reserveMux.Lock()
	defer r.reserveMux.Unlock()

	id, ok := m.Query["id"]
	if !ok || id == "" {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", id), read, write)
	}

	seats, ok := m.Query["seats"]
	if !ok || id == "" {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", id), read, write)
	}

	// Convert id to int
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", idInt), read, write)
	}

	// Convert seats to int
	seatsInt, err := strconv.ParseInt(seats, 10, 64)
	if err != nil || seatsInt <= 0 {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", seatsInt), read, write)
	}
	seatsInt32 := int32(seatsInt)

	// Find flight by id from repo
	flight, err := r.flightRepo.FindByID(int32(idInt))
	if err != nil {
		return r.error(method, err, "", read, write)
	}
	if flight == nil {
		return r.error(method, ErrNoFlightFound, fmt.Sprintf("No flights found with %v", idInt), read, write)
	}

	// Check if flight has enough seats to reserve
	if flight.SeatAvailablity-seatsInt32 < 0 {
		return r.error(method, ErrFailToReserve, fmt.Sprintf("Not enough seats to reserve for flight %v", idInt), read, write)
	}

	// Update flight seat availability
	flight.SeatAvailablity -= seatsInt32

	// Update flight in repo
	if err = r.flightRepo.Update(flight); err != nil {
		return r.error(method, ErrFailToReserve, err.Error(), read, write)
	}

	reserve := &ReserveFlight{
		ID:           uuid.NewString(),
		Flight:       flight,
		SeatReserved: seatsInt32,
		CheckIn:      false,
	}

	// Create a reservation in repo
	if err = r.reservationRepo.Insert(reserve); err != nil {
		return r.error(method, ErrFailToReserve, err.Error(), read, write)
	}

	// Broadcast flight updates to listening channels
	r.broadcastFlights(flight)

	b, err := encoding.Marshal(reserve)
	if err != nil {
		return r.error(method, ErrInternalError, err.Error(), read, write)
	}

	return r.ok(method, b, lossy, read, write)
}

// CheckInFlight checks in a flight by `id` in query params. This is an idempotent method
func (r *RPC) CheckInFlight(m *Message, read Readable, write Writable) error {
	method := "CheckInFlight"
	lossy := true

	r.reserveMux.Lock()
	defer r.reserveMux.Unlock()

	id, ok := m.Query["id"]
	if !ok || id == "" {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", id), read, write)
	}

	// Retrieve reservation
	rf, err := r.reservationRepo.FindByID(id)
	if err != nil {
		return r.error(method, ErrNoReserveFlightFound, "", read, write)
	}

	// Marshal function to return response
	marshal := func(reserve *ReserveFlight) error {
		b, err := encoding.Marshal(rf)
		if err != nil {
			return r.error(method, ErrInternalError, err.Error(), read, write)
		}
		return r.ok(method, b, lossy, read, write)
	}

	// returns if already checked in
	if rf.CheckIn {
		return marshal(rf)
	}

	// Update reservation
	rf.CheckIn = true
	if err = r.reservationRepo.Update(rf); err != nil {
		return r.error(method, ErrInternalError, err.Error(), read, write)
	}
	return marshal(rf)
}

// GetMeals returns a list of meals. This is an idempotent method
func (r *RPC) GetMeals(read Readable, write Writable) error {
	method := "GetMeals"
	meals := GetFood()
	lossy := true

	// convert meals to list
	mealList := []*Food{}
	for _, meal := range meals {
		mealList = append(mealList, meal)
	}

	b, err := encoding.Marshal(mealList)
	if err != nil {
		return r.error(method, ErrInternalError, err.Error(), read, write)
	}

	return r.ok(method, b, lossy, read, write)
}

// AddMeals adds a meal to the list of meals. This is a non-idempotent method
func (r *RPC) AddMeals(m *Message, read Readable, write Writable) error {
	method := "AddMeals"
	lossy := true

	r.reserveMux.Lock()
	defer r.reserveMux.Unlock()

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
	meal, ok := meals[int32(mealId)]
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
	return r.ok(method, b, lossy, read, write)
}

// CancelFlight cancels a flight by `id` in query params. This is an idempotent method
func (r *RPC) CancelFlight(m *Message, read Readable, write Writable) error {
	method := "CancelFlight"
	lossy := true

	res := func(rf *ReserveFlight) error {
		b, err := encoding.Marshal(rf)
		if err != nil {
			return r.error(method, ErrInternalError, err.Error(), read, write)
		}
		return r.ok(method, b, lossy, read, write)
	}

	r.reserveMux.Lock()
	defer r.reserveMux.Unlock()

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

	// Return if already cancelled
	if rf.Cancelled {
		return res(rf)
	}

	// Prevent cancellation for checked in flights
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

	rf.Cancelled = true
	// Update reservation
	err = r.reservationRepo.Update(rf)
	if err != nil {
		return r.error(method, ErrInternalError, err.Error(), read, write)
	}

	// Update seats availability
	flight.SeatAvailablity += rf.SeatReserved
	if err = r.flightRepo.Update(flight); err != nil {
		return r.error(method, ErrInternalError, err.Error(), read, write)
	}

	r.broadcastFlights(flight)

	return res(rf)
}

// MonitorUpdates monitors flight updates for a given timestamp deadline in query params. This is a non-idempotent method
// Caveat for this method is that if a client disconnect, it will not immediately delete the channel
func (r *RPC) MonitorUpdates(addr string, m *Message, read Readable, write Writable) error {
	method := "MonitorUpdates"
	// We ensure MonitorUpdates is not susceptible to frame drops
	lossy := false
	t, ok := m.Query["timestamp"]
	if !ok || t == "" {
		return r.error(method, ErrInvalidParams, fmt.Sprintf("%v: is invalid", t), read, write)
	}

	// convert string unix timestamp to time.Time
	monitorUntil, err := common.StrToUnixTime(t)
	if err != nil {
		return r.error(method, ErrInvalidParams, err.Error(), read, write)
	}

	// Create channel
	if _, ok := r.chFlightUpdates[addr]; ok {
		return r.error(method, ErrInternalError, "Already monitoring", read, write)
	}
	r.chFlightUpdatesMux.Lock()
	fCh := make(chan *Flight)
	r.chFlightUpdates[addr] = fCh
	r.chFlightUpdatesMux.Unlock()

	r.logger.Info("Current Time : ", time.Now().Local().Format(time.RFC3339))
	r.logger.Info("Deadline     : ", monitorUntil.Local().Format(time.RFC3339))

	duration := time.Until(*monitorUntil)
	// Listen for updates
	for {
		select {
		// Listen for updates via flight channel
		case flight := <-fCh:
			// Updates client
			b, err := encoding.Marshal(flight)
			if err != nil {
				return r.error(method, err, "", read, write)
			}
			r.ok(method, b, lossy, read, write)

		// Listens for deadline
		case <-time.After(duration):
			// Remove channel
			r.chFlightUpdatesMux.Lock()
			// Remove via index
			delete(r.chFlightUpdates, addr)
			r.chFlightUpdatesMux.Unlock()
			r.logger.Info(fmt.Sprintf("Released: %v", addr))
			return r.ok(method, []byte{}, lossy, read, write)

		}
	}
}

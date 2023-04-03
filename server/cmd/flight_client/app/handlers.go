package app

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/isaiahwong/cz4013/common"
	"github.com/manifoldco/promptui"
)

// FindFlights facade function that handle user prompts
// and delegates to client for actual RPC calls
func (a *App) FindFlights() {
	srcP := promptui.Prompt{
		Label: "Enter source",
	}

	destP := promptui.Prompt{
		Label: "Enter destination",
	}

	// Prompt for source
	src, err := srcP.Run()
	if err != nil {
		a.logger.WithError(err).Error(FindFlights)
		return
	}

	// Prompt for destination
	dest, err := destP.Run()
	if err != nil {
		a.logger.WithError(err).Error(FindFlights)
		return
	}

	// Call RPC Method
	flights, err := a.c.FindFlights(src, dest)
	if err != nil {
		a.logger.WithError(err).Error(FindFlights)
		return
	}

	if len(flights) == 0 {
		common.PrintTitle("No flights found")
		return
	}

	common.PrintTitle("Flights")
	// Print for equal alignment
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 4, ' ', tabwriter.TabIndent)
	for i, f := range flights {
		fmt.Println(fmt.Sprintf("\nFlight %v", i+1))
		fmt.Fprintln(w, f.String())
	}
	// Flush the tabwriter to write the output
	w.Flush()
}

// FindFlight facade function that handle user prompts
// and delegates to client for actual RPC calls
func (a *App) FindFlight() {
	idP := promptui.Prompt{
		Label: "Enter flight id",
	}

	id, err := idP.Run()
	if err != nil {
		a.logger.WithError(err).Error(FindFlight)
		return
	}

	// Call RPC Method
	flight, err := a.c.FindFlight(id)
	if err != nil {
		a.logger.WithError(err).Error(FindFlight)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 4, ' ', tabwriter.TabIndent)
	common.PrintTitle("Flight")
	fmt.Fprintln(w, flight.String())
	// Flush the tabwriter to write the output
	w.Flush()
}

// ReserveFlight facade function that handle user prompts
// and delegates to client for actual RPC calls
func (a *App) ReserveFlight() {
	f := promptui.Prompt{
		Label:    "Enter Flight ID",
		Validate: common.ValidateInt,
	}

	s := promptui.Prompt{
		Label:    "Enter seats",
		Validate: common.ValidateInt,
	}

	flightID, err := f.Run()
	if err != nil {
		a.logger.WithError(err).Error(ReserveFlight)
		return
	}

	seatsStr, err := s.Run()
	if err != nil {
		a.logger.WithError(err).Error(ReserveFlight)
		return
	}
	// Assumes validator is correct
	seats, _ := strconv.ParseInt(seatsStr, 10, 32)

	reservation, err := a.c.ReserveFlight(flightID, int(seats))
	if err != nil {
		a.logger.WithError(err).Error(ReserveFlight)
		return
	}

	// Save reservation
	a.c.Reservations[reservation.ID] = reservation

	// print reservation
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 4, ' ', tabwriter.TabIndent)

	common.PrintTitle("Reservation Successful")
	fmt.Println("\nReservation Details")
	fmt.Fprintln(w, reservation.String())

	fmt.Println("Flight Details")
	fmt.Fprintln(w, reservation.Flight.String())

	// Flush the tabwriter to write the output
	w.Flush()
}

// CheckInFlight facade function that handle user prompts
// and delegates to client for actual RPC calls
func (a *App) CheckInFlight() {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 4, ' ', tabwriter.TabIndent)

	// Index to reservation id
	indexToID := make(map[int]string)
	i := 0

	for k, v := range a.c.Reservations {
		indexToID[i] = k
		fmt.Println(fmt.Sprintf("\nReservation [%v]", i))
		fmt.Fprintln(w, v.String())
		i++
	}
	w.Flush()

	rp := promptui.Prompt{
		Label:    "Select index of reservation cancel",
		Validate: common.ValidateRange(0, int64(len(a.c.Reservations))),
	}
	reserveStr, err := rp.Run()
	if err != nil {
		a.logger.WithError(err).Error(CancelFlight)
		return
	}

	reserveIdx, _ := strconv.ParseInt(reserveStr, 10, 32)
	reservationID, ok := indexToID[int(reserveIdx)]
	if !ok {
		a.logger.Error("Reservation doesn't exists!")
		return
	}

	flight, err := a.c.CheckInFlight(reservationID)
	if err != nil {
		a.logger.WithError(err).Error(CancelFlight)
		return
	}

	common.PrintTitle("Reservation Cancelled")
	fmt.Fprintln(w, flight.String())
	// Flush the tabwriter to write the output
	w.Flush()
}

// AddMeals facade function that handle user prompts
// and delegates to client for actual RPC calls
func (a *App) AddMeals() {
	if len(a.c.Reservations) == 0 {
		a.logger.Info("No reservations made")
		return
	}

	// Used for equal column printing
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 4, ' ', tabwriter.TabIndent)

	// Retrieve meals from server
	meals, err := a.c.GetMeals()
	if err != nil {
		a.logger.WithError(err).Error(GetMeals)
		return
	}

	if len(meals) == 0 {
		return
	}

	// Index to reservation id
	indexToID := make(map[int]string)
	i := 0

	// Print reservation for selection
	for k, v := range a.c.Reservations {
		indexToID[i] = k
		fmt.Println(fmt.Sprintf("Reservation [%v]", i))
		fmt.Fprintln(w, v.String())
		i++
	}
	w.Flush()

	// Prompt users to select reservation
	rp := promptui.Prompt{
		Label:    "Select index of reservation to add meal",
		Validate: common.ValidateRange(0, int64(len(a.c.Reservations))),
	}
	reserveStr, err := rp.Run()
	if err != nil {
		a.logger.WithError(err).Error(AddMeals)
		return
	}
	reserveIdx, _ := strconv.ParseInt(reserveStr, 10, 32)
	reservationID, ok := indexToID[int(reserveIdx)]
	if !ok {
		a.logger.Error("Reservation doesn't exists!")
		return
	}

	// Print meals for selection
	for i, v := range meals {
		fmt.Println(fmt.Sprintf("Meal [%v]", i))
		fmt.Fprintln(w, v.String())
	}
	w.Flush()

	mp := promptui.Prompt{
		Label:    "Select index of meal to add",
		Validate: common.ValidateRange(0, int64(len(meals))),
	}
	mealStr, err := mp.Run()
	if err != nil {
		a.logger.WithError(err).Error(AddMeals)
		return
	}
	mealIdx, _ := strconv.ParseInt(mealStr, 10, 32)
	meal := meals[int(mealIdx)]

	// Add meal to reservation RPC call
	rc, err := a.c.AddMeals(reservationID, fmt.Sprint(meal.ID))
	if err != nil {
		a.logger.WithError(err).Error(AddMeals)
		return
	}

	// Update local state reservation
	a.c.Reservations[rc.ID] = rc
	common.PrintTitle("Meal Added")
	fmt.Fprintln(w, rc.String())
	fmt.Fprintln(w, "Meal Details")
	for _, meal := range rc.Meals {
		fmt.Fprintln(w, meal.String())
	}
	// Flush the tabwriter to write the output
	w.Flush()
}

// CancelFlight facade function that handle user prompts
// and delegates to client for actual RPC calls
func (a *App) CancelFlight() {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 4, ' ', tabwriter.TabIndent)

	// Index to reservation id
	indexToID := make(map[int]string)
	i := 0

	// Print reservation for selection
	for k, v := range a.c.Reservations {
		indexToID[i] = k
		fmt.Println(fmt.Sprintf("\nReservation [%v]", i))
		fmt.Fprintln(w, v.String())
		i++
	}
	w.Flush()

	rp := promptui.Prompt{
		Label:    "Select index of reservation cancel",
		Validate: common.ValidateRange(0, int64(len(a.c.Reservations))),
	}
	reserveStr, err := rp.Run()
	if err != nil {
		a.logger.WithError(err).Error(CancelFlight)
		return
	}

	reserveIdx, _ := strconv.ParseInt(reserveStr, 10, 32)

	reservationID, ok := indexToID[int(reserveIdx)]
	if !ok {
		a.logger.Error("Reservation doesn't exists!")
		return
	}

	reserveFlight, err := a.c.CancelFlight(reservationID)
	if err != nil {
		a.logger.WithError(err).Error(CancelFlight)
		return
	}

	common.PrintTitle("Reservation Cancelled")
	fmt.Fprintln(w, reserveFlight.String())
	// Flush the tabwriter to write the output
	w.Flush()
}

// MonitorUpdates facade function that handle user prompts
// Function handles keyStroke access for user to cancel monitoring
func (a *App) MonitorUpdates() {
	// Validate inline function
	validate := func(input string) error {
		i, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return errors.New("Invalid input")
		}

		if i <= 0 {
			return errors.New("Must be greater than 0")
		}
		return nil
	}

	f := promptui.Prompt{
		Label:    "Enter Flight ID",
		Validate: common.ValidateInt,
	}
	flightID, err := f.Run()
	if err != nil {
		a.logger.WithError(err).Error(ReserveFlight)
		return
	}

	ti := promptui.Prompt{
		Label:    "Enter how long to monitor for (in minutes)",
		Validate: validate,
	}
	ts, err := ti.Run()
	if err != nil {
		a.logger.WithError(err).Error(MonitorUpdates)
		return
	}
	// Convert to minutes
	t, _ := strconv.ParseInt(ts, 10, 32)
	if err != nil {
		a.logger.WithError(err).Error(MonitorUpdates)
		return
	}

	// Listen for keystroke
	go a.onKeyStoke()

	// Perform block RPC calls
	err = a.c.MonitorUpdates(flightID, time.Duration(t)*time.Minute, a.keyStrokeCh)
}

func (a *App) ViewReservations() {
	if len(a.c.Reservations) == 0 {
		a.logger.Info("No reservations made")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 4, ' ', tabwriter.TabIndent)
	common.PrintTitle("Reservations")
	for _, v := range a.c.Reservations {
		fmt.Fprintln(w, v.String())
		fmt.Fprintln(w, "Meal Details")
		for _, meal := range v.Meals {
			fmt.Fprintln(w, meal.String())
		}
	}
	// Flush the tabwriter to write the output
	w.Flush()

}

package app

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/isaiahwong/cz4013/cmd/flight_client/client"
	"github.com/manifoldco/promptui"
)

var (
	FindFlights    = "FindFlights"
	FindFlight     = "FindFlight"
	ReserveFlight  = "ReserveFlight"
	CheckInFlight  = "CheckInFlight"
	CancelFlight   = "CancelFlight"
	MonitorUpdates = "MonitorUpdates"
)

var validateInt = func(input string) error {
	_, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return errors.New("Invalid input")
	}
	return nil
}

type App struct {
	c       *client.Client
	options promptui.Select
}

func (a *App) reserveFlight() {
	f := promptui.Prompt{
		Label:    "Enter Flight ID",
		Validate: validateInt,
	}

	s := promptui.Prompt{
		Label:    "Enter seats",
		Validate: validateInt,
	}

	flightID, err := f.Run()
	if err != nil {
		fmt.Println(err)
		return
	}

	seatsStr, err := s.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
	seats, _ := strconv.ParseInt(seatsStr, 10, 32)

	reservation, err := a.c.ReserveFlight(flightID, int(seats))
	if err != nil {
		fmt.Println(err)
		return
	}

	// print reservation
	fmt.Println("Reservation Details:")
	fmt.Println(reservation.String())
	fmt.Println("\nFlight Details:")
	fmt.Println(reservation.Flight.String())
}

func (a *App) topLevel() {
	_, result, err := a.options.Run()

	if err == promptui.ErrInterrupt {
		fmt.Println("Goodbye")
		os.Exit(0)
	}

	if err != nil {
		fmt.Println(err)
		return
	}

	switch result {
	case FindFlights:
	case FindFlight:
	case ReserveFlight:
		a.reserveFlight()
	case CheckInFlight:
	case CancelFlight:
	case MonitorUpdates:
	}
}

func (a *App) Start() error {
	err := a.c.Start()
	if err != nil {
		return err
	}
	for {
		a.topLevel()
	}
}

func New(c *client.Client) *App {
	options := promptui.Select{
		Label: "Select an RPC method",
		Items: []string{
			FindFlights, FindFlight,
			ReserveFlight, CheckInFlight,
			CancelFlight, MonitorUpdates,
		},
		Size: 10,
	}

	return &App{
		c:       c,
		options: options,
	}
}

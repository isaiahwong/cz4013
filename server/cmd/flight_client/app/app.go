package app

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/isaiahwong/cz4013/cmd/flight_client/client"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
)

var (
	FindFlights      = "FindFlights"
	FindFlight       = "FindFlight"
	ReserveFlight    = "ReserveFlight"
	CheckInFlight    = "CheckInFlight"
	CancelFlight     = "CancelFlight"
	ViewReservations = "ViewReservations"
	GetMeals         = "GetMeals"
	AddMeals         = "AddMeals"
	MonitorUpdates   = "MonitorUpdates"
)

var validateInt = func(input string) error {
	_, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return errors.New("Invalid input")
	}
	return nil
}

var validateIndex = func(l, r int64) func(input string) error {
	return func(input string) error {
		in, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return errors.New("Invalid input")
		}

		if in < l || in > r {
			return errors.New("Selection out of range")
		}
		return nil
	}

}

type App struct {
	c           *client.Client
	options     promptui.Select
	logger      *logrus.Logger
	keyStrokeCh chan struct{}
}

func (a *App) printTitle(title string) {
	fmt.Println("========================================")
	fmt.Println(title)
	fmt.Println("========================================")
}

func (a *App) onKeyStoke() {
	fmt.Println("Press enter to cancel...")
	reader := bufio.NewReader(os.Stdin)
	_, _, _ = reader.ReadRune()
	select {
	case a.keyStrokeCh <- struct{}{}:
	}
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
		a.findFlights()
	case FindFlight:
		a.findFlight()
	case ReserveFlight:
		a.reserveFlight()
	case CheckInFlight:
		a.checkInFlight()
	case CancelFlight:
		a.cancelFlight()
	case ViewReservations:
		a.ViewReservations()
	case AddMeals:
		a.AddMeals()
	case MonitorUpdates:
		a.monitorUpdates()
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
			ViewReservations,
			AddMeals, MonitorUpdates,
		},
		Size: 10,
	}

	return &App{
		c:           c,
		options:     options,
		logger:      logrus.New(),
		keyStrokeCh: make(chan struct{}, 1),
	}
}

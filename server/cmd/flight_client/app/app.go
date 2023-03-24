package app

import (
	"bufio"
	"fmt"
	"os"
	"time"

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
	Exit             = "Exit"
)

type App struct {
	c           *client.Client
	options     promptui.Select
	logger      *logrus.Logger
	keyStrokeCh chan client.Placeholder
}

func (a *App) onKeyStoke() {
	fmt.Println("Press enter to cancel...")
	reader := bufio.NewReader(os.Stdin)
	_, _, _ = reader.ReadRune()
	select {
	case a.keyStrokeCh <- client.Placeholder{Time: time.Now()}:
	}

}

func (a *App) topLevel() {
	close := func() {
		fmt.Println("Goodbye")
		os.Exit(0)
	}
	_, result, err := a.options.Run()

	if err == promptui.ErrInterrupt {
		close()
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
	case Exit:
		close()
		return
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
			ReserveFlight, CancelFlight,
			ViewReservations,
			AddMeals, MonitorUpdates, Exit,
		},
		Size: 10,
	}

	return &App{
		c:           c,
		options:     options,
		logger:      logrus.New(),
		keyStrokeCh: make(chan client.Placeholder),
	}
}

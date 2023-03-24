package server

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/isaiahwong/cz4013/common"
	"github.com/isaiahwong/cz4013/protocol"
	"github.com/isaiahwong/cz4013/rpc"
	"github.com/isaiahwong/cz4013/store"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
)

var db *store.DB
var flightRepo *rpc.FlightRepo
var reservationRepo *rpc.ReservationRepo
var logger *logrus.Logger

func init() {
	var flights = []*rpc.Flight{}

	// Load flights from csv
	if err := common.LoadCSV("flights.csv", &flights); err != nil {
		panic(err)
	}
	// Create a new db
	db = store.New()
	fType := new(rpc.Flight)
	rType := new(rpc.ReserveFlight)
	// Create new data repositories
	flightRepo = rpc.NewFlightRepo(db)
	reservationRepo = rpc.NewReservationRepo(db)

	// Create respective relations
	if err := db.CreateRelation(flightRepo.Relation, reflect.TypeOf(fType)); err != nil {
		panic(err)
	}
	if err := db.CreateRelation(reservationRepo.Relation, reflect.TypeOf(rType)); err != nil {
		panic(err)
	}
	if err := db.BulkInsert(flightRepo.Relation, flights); err != nil {
		panic(err)
	}

	logger = logrus.New()
}

// New creates a new server with the specified parameters
func New(semantic int, deadline int, lossRate int, port string) *protocol.Server {
	return protocol.New(
		protocol.WithSemantic(protocol.IntToSemantics(semantic)),
		protocol.WithDeadline(time.Duration(deadline)*time.Second),
		protocol.WithFlightRepo(flightRepo),
		protocol.WithReservationRepo(reservationRepo),
		protocol.WithLossRate(lossRate),
		protocol.WithLogger(logger),
		protocol.WithPort(fmt.Sprintf(":%v", port)),
	)
}

// handleInterrupt handles the interrupt keyboard interrupts
func handleInterrupt(err error) error {
	if err == promptui.ErrInterrupt {
		fmt.Println("Goodbye")
		os.Exit(0)
	}
	return err
}

// prompt provides an interactive prompt to configure the server
func prompt() *protocol.Server {
	loadDefault := "Load default config"
	customConfig := "Custom config"
	semantics := []protocol.Semantics{
		protocol.AtMostOnce,
		protocol.AtLeastOnce,
	}

	sp := promptui.Select{
		Label: "Select option",
		Items: []string{
			loadDefault,
			customConfig,
		},
	}
	_, input, err := sp.Run()
	if common.HandleInterrupt(err) != nil {
		panic(err)
	}

	if input == loadDefault {
		return New(0, 5, 0, "8080")
	}

	// Custom config
	semP := promptui.Select{
		Label: "Select semantics",
		Items: semantics,
	}

	lossRate := promptui.Prompt{
		Label:    "Enter loss rate 0 - 100",
		Validate: common.ValidateRange(0, 100),
	}

	semIdx, _, err := semP.Run()
	if common.HandleInterrupt(err) != nil {
		panic(err)
	}

	lossRateInput, err := lossRate.Run()
	if common.HandleInterrupt(err) != nil {
		panic(err)
	}
	lossRateInt, _ := strconv.ParseInt(lossRateInput, 10, 32)

	return New(semIdx, 5, int(lossRateInt), "8080")
}

// Start starts the server
func Start() {
	s := prompt()
	s.Serve()
}

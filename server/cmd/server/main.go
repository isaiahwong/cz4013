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
)

var flights = []*rpc.Flight{}
var db *store.DB
var flightRepo *rpc.FlightRepo
var reservationRepo *rpc.ReservationRepo

func init() {
	// Load flights from csv
	if err := common.LoadCSV("flights.csv", &flights); err != nil {
		panic(err)
	}
	db = store.New()
	fType := new(rpc.Flight)
	rType := new(rpc.ReserveFlight)

	flightRepo = rpc.NewFlightRepo(db)
	reservationRepo = rpc.NewReservationRepo(db)

	if err := db.CreateRelation(flightRepo.Relation, reflect.TypeOf(fType)); err != nil {
		panic(err)
	}
	if err := db.CreateRelation(reservationRepo.Relation, reflect.TypeOf(rType)); err != nil {
		panic(err)
	}
	if err := db.BulkInsert(flightRepo.Relation, flights); err != nil {
		panic(err)
	}
}

func handleInterrupt(err error) error {
	if err == promptui.ErrInterrupt {
		fmt.Println("Goodbye")
		os.Exit(0)
	}
	return err
}

func prompt() *protocol.Server {
	loadDefault := "Load default config"
	customConfig := "Custom config"
	semantics := []*protocol.Semantic{
		{Name: "AtLeastOnce", Value: protocol.AtLeastOnce},
		{Name: "AtMostOnce", Value: protocol.AtMostOnce},
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
		return protocol.New(
			protocol.WithSemantic(protocol.AtMostOnce),
			protocol.WithDeadline(5*time.Second),
			protocol.WithFlightRepo(flightRepo),
			protocol.WithReservationRepo(reservationRepo),
			protocol.WithLossRate(0),
		)
	}

	// Custom config
	semP := promptui.Select{
		Label: "Select semantics",
		Items: []*protocol.Semantic{
			{Name: "AtLeastOnce", Value: protocol.AtLeastOnce},
			{Name: "AtMostOnce", Value: protocol.AtMostOnce},
		},
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

	return protocol.New(
		protocol.WithSemantic(semantics[semIdx].Value),
		protocol.WithDeadline(5*time.Second),
		protocol.WithFlightRepo(flightRepo),
		protocol.WithReservationRepo(reservationRepo),
		protocol.WithLossRate(int(lossRateInt)),
	)
}

func Start() {
	s := prompt()
	s.Serve()
}

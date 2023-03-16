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

var flights = []*rpc.Flight{}
var db *store.DB
var flightRepo *rpc.FlightRepo
var reservationRepo *rpc.ReservationRepo
var logger *logrus.Logger

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

	logger = logrus.New()
	// logrus.Trace()
	// file, err := os.OpenFile("./logs/server.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	// if err != nil {
	// 	logger.Fatal(err)
	// }
	// // Create a multi-writer that writes to both the console and the file
	// writer := io.MultiWriter(os.Stdout, file)

	// // Set the logger's output to the multi-writer
	// logger.SetOutput(writer)
}

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

func Start() {
	s := prompt()
	s.Serve()
}

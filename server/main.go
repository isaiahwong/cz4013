package main

import (
	"reflect"
	"time"

	"github.com/isaiahwong/cz4013/common"
	"github.com/isaiahwong/cz4013/protocol"
	"github.com/isaiahwong/cz4013/rpc"
	"github.com/isaiahwong/cz4013/store"
)

// func main() {
// 	f := &rpc.Flight{Source: "asdas"}
// 	b, err := encoding.Marshal(f)
// 	if err != nil {
// 		panic(err)
// 	}
// 	f2 := new(rpc.Flight)
// 	err = encoding.Unmarshal(b, f2)
// 	if err != nil {
// 		panic(err)
// 	}
// 	print(f2.Source)
// }

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

func main() {
	s := protocol.New(
		protocol.WithSemantic(protocol.AtMostOnce),
		protocol.WithDeadline(5*time.Second),
		protocol.WithFlightRepo(flightRepo),
		protocol.WithReservationRepo(reservationRepo),
	)
	s.Serve()
}

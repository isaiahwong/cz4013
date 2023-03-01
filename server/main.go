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

func init() {
	// Load flights from csv
	if err := common.LoadCSV("flights.csv", &flights); err != nil {
		panic(err)
	}
	db = store.New()
	fType := new(rpc.Flight)
	if err := db.CreateRelation("flights", reflect.TypeOf(fType)); err != nil {
		panic(err)
	}
	if err := db.BulkInsert("flights", flights); err != nil {
		panic(err)
	}

	flightRepo = rpc.NewFlightRepo(db)
}

func main() {
	s := protocol.New(
		protocol.WithDeadline(5*time.Second),
		protocol.WithFlightRepo(flightRepo),
	)
	s.Serve()
}

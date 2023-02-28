package main

import (
	"reflect"
	"time"

	"github.com/isaiahwong/cz4013/common"
	"github.com/isaiahwong/cz4013/protocol"
	"github.com/isaiahwong/cz4013/rpc"
	"github.com/isaiahwong/cz4013/store"
)

var flights = []rpc.Flight{}
var db *store.DB

func init() {
	// Load flights from csv
	common.LoadCSV("flights.csv", &flights)
	db = store.New()

	fType := new(rpc.Flight)
	db.CreateRelation("flights", reflect.TypeOf(fType))
	db.BulkInsert("flights", flights)
}

func main() {
	s := protocol.New(
		protocol.WithDeadline(5*time.Second),
		protocol.WithDB(db),
	)
	s.Serve()
}

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

// package main

// import (
// 	"fmt"
// 	"net"
// 	"time"
// )

// const serverPort = ":8080"

// func main() {
// 	fmt.Println("A  Basic UDP Server Example")

// 	ServerAddr, err := net.ResolveUDPAddr("udp", serverPort)
// 	if err != nil {
// 		panic(err)
// 	}

// 	ServerConn, err := net.ListenUDP("udp", ServerAddr)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer ServerConn.Close()

// 	buf := make([]byte, 1024)
// 	go func() {
// 		for {
// 			n, addr, err := ServerConn.ReadFromUDP(buf)
// 			fmt.Println("Received ", string(buf[0:n]), " from ", addr)

// 			if err != nil {
// 				fmt.Println("Error: ", err)
// 			}
// 			time.Sleep(time.Second * 5)
// 			//after we got something, respond with an "OK" to the client
// 			buf = []byte("OK")
// 			ServerConn.WriteToUDP(buf, addr)
// 		}
// 	}()

// 	fmt.Println("Waiting for clients to connect. Server port " + serverPort)

// 	//blocking forever
// 	select {}
// }

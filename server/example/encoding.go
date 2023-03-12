package main

import (
	"fmt"
	"time"

	"github.com/isaiahwong/cz4013/encoding"
	"github.com/isaiahwong/cz4013/rpc"
)

func main() {
	// Create random flight with values
	flight := rpc.Flight{
		ID:              1,
		Source:          "SIN",
		Destination:     "JFK",
		Airfare:         100.0,
		SeatAvailablity: 100,
		Timestamp:       uint32(time.Now().Unix()),
	}

	// Marshal the flight
	buf, _ := encoding.Marshal(flight)
	fmt.Println(buf)
}

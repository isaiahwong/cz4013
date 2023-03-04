package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/isaiahwong/cz4013/encoding"
	"github.com/isaiahwong/cz4013/protocol"
	"github.com/isaiahwong/cz4013/rpc"
)

func monitorUpdates() {
	// Create a UDP address for the server
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	// Create a UDP connection to the server
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		panic(err)
	}

	fmt.Println(conn.LocalAddr())

	session := protocol.NewSession(conn, true)
	session.Start()
	// Open a new stream
	stream, err := session.Open(serverAddr)
	if err != nil {
		panic(err)
	}

	v := &rpc.Message{
		RPC: "MonitorUpdates",
		Query: map[string]string{
			"timestamp": fmt.Sprintf("%v", time.Now().Add(10000*time.Second).Unix()*1000),
		},
		Body: []byte{},
	}
	b, err := encoding.Marshal(v)
	if err != nil {
		panic(err)
	}

	stream.Write(b)

	for !stream.IsClosed() {
		res := make([]byte, 65507)
		// stream.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := stream.Read(res)
		if err != nil && err != io.EOF {
			panic(err)
		}

		m := new(rpc.Message)
		flight := new(rpc.Flight)

		if err = encoding.Unmarshal(res[:n], m); err != nil && err != io.EOF {
			panic(err)
		}

		if err = encoding.Unmarshal(m.Body, flight); err != nil && err != io.EOF {
			panic(err)
		}
		fmt.Println("New Updated flight: ", flight)
	}

	stream.Close()
	session.Close()
}

func reserveFlight() {

	// Create a UDP address for the server
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	// Create a UDP connection to the server
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		panic(err)
	}

	session := protocol.NewSession(conn, true)
	session.Start()

	for {
		// Open a new stream
		stream, err := session.Open(serverAddr)
		if err != nil {
			panic(err)
		}
		v := &rpc.Message{
			RPC: "ReserveFlight",
			Query: map[string]string{
				"id":    "5653",
				"seats": "1",
			},
			Body: []byte{},
		}
		b, err := encoding.Marshal(v)
		if err != nil {
			panic(err)
		}

		stream.Write(b)
		res := make([]byte, 65507)
		stream.SetReadDeadline(time.Now().Add(5 * time.Second))

		n, err := stream.Read(res)
		if err != nil && err != io.EOF {
			panic(err)
		}

		m := new(rpc.Message)
		reservation := new(rpc.ReserveFlight)
		if err = encoding.Unmarshal(res[:n], m); err != nil && err != io.EOF {
			panic(err)
		}

		if m.Error != nil {
			fmt.Printf("Error reserving: %v\n", m.Error)
			return
		}
		if err = encoding.Unmarshal(m.Body, reservation); err != nil && err != io.EOF {
			panic(err)
		}
		stream.Close()

		stream, err = session.Open(serverAddr)
		if err != nil {
			panic(err)
		}
		cancelMsg := &rpc.Message{
			RPC: "CancelFlight",
			Query: map[string]string{
				"id": reservation.ID,
			},
			Body: []byte{},
		}
		b, err = encoding.Marshal(cancelMsg)
		if err != nil {
			panic(err)
		}

		stream.Write(b)
		stream.SetReadDeadline(time.Now().Add(5 * time.Second))

		n, err = stream.Read(res)
		if err != nil && err != io.EOF {
			panic(err)
		}

		m = new(rpc.Message)
		flight := new(rpc.Flight)

		if err = encoding.Unmarshal(res[:n], m); err != nil && err != io.EOF {
			panic(err)
		}

		if m.Error != nil {
			fmt.Printf("Error Cancelling: %v\n", m.Error)
			return
		}
		if err = encoding.Unmarshal(m.Body, flight); err != nil && err != io.EOF {
			panic(err)
		}
		stream.Close()
		fmt.Println("Cancelled successfully")
		time.Sleep(1 * time.Second)
	}

}

func main() {
	go reserveFlight()
	monitorUpdates()
}

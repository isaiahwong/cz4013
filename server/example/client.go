package main

import (
	"fmt"
	"net"
	"time"

	"github.com/isaiahwong/cz4013/encoding"
	"github.com/isaiahwong/cz4013/protocol"
	"github.com/isaiahwong/cz4013/rpc"
)

func main() {

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
	for {
		session.Start()
		// Open a new stream
		stream, err := session.Open(serverAddr)
		if err != nil {
			panic(err)
		}

		v := rpc.Message{
			RPC:   "Create",
			Query: map[string]string{"key": "value"},
			Body:  []byte{}}
		b, err := encoding.Marshal(v)
		if err != nil {
			panic(err)
		}

		stream.Write(b)

		res := make([]byte, 1024)
		// stream.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := stream.Read(res)
		if err != nil {
			panic(err)
		}

		var m rpc.Message
		_ = encoding.Unmarshal(res[:n], &m)
		fmt.Println(m.Error.Body)
		if err != nil {
			panic(err)
		}

		stream.Close()
		time.Sleep(time.Second * 1)
		// session.Close()
	}

}

package main

import (
	"fmt"
	"net"
	"time"

	"github.com/isaiahwong/cz4013/encoding"
	"github.com/isaiahwong/cz4013/protocol"
	"github.com/isaiahwong/cz4013/rpc"
)

const chunkSize = 1024

var sent = 0

func splitMessage(msg []byte) [][]byte {
	var chunks [][]byte

	// Calculate the number of chunks needed to split the message
	numChunks := len(msg) / chunkSize
	if len(msg)%chunkSize != 0 {
		numChunks++
	}

	// Split the message into chunks
	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(msg) {
			end = len(msg)
		}
		chunk := msg[start:end]
		chunks = append(chunks, chunk)
	}

	return chunks
}

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

	session := protocol.NewSession(conn)
	for {
		session.Start()
		// Open a new stream
		stream, err := session.Open(serverAddr)
		if err != nil {
			panic(err)
		}

		// stream.Close()
		// time.Sleep(1 * time.Hour)

		p := rpc.Person{
			Name:    "John",
			Friends: []*rpc.Person{{Name: "Bob"}, {Name: "Alice"}},
		}

		pb, err := encoding.Marshal(p)
		if err != nil {
			panic(err)
		}

		v := rpc.Message{Sent: int32(sent), RPC: "Create", Body: pb}
		b, err := encoding.Marshal(v)
		if err != nil {
			panic(err)
		}

		var m rpc.Message
		_ = encoding.Unmarshal(b, &m)
		if err != nil {
			panic(err)
		}
		fmt.Println(b)

		fmt.Println(len(b))
		sent++
		stream.Write(b)
		fmt.Println("Sent", sent)
		time.Sleep(100 * time.Millisecond)
		stream.Close()
		// session.Close()
	}

}

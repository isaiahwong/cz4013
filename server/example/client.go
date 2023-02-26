package main

import (
	"fmt"
	"net"

	"github.com/isaiahwong/cz4013/encoding"
	"github.com/isaiahwong/cz4013/protocol"
	"github.com/isaiahwong/cz4013/rpc"
)

const chunkSize = 1024

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
	session.Start()
	// Open a new stream
	stream, err := session.Open(serverAddr)
	if err != nil {
		panic(err)
	}

	p := rpc.Person{
		Name:    "John",
		Friends: []*rpc.Person{{Name: "Bob"}, {Name: "Alice"}},
	}

	pb, err := encoding.Marshal(p)
	if err != nil {
		panic(err)
	}

	v := rpc.Message{RPC: "Create", Body: pb}
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
	stream.Write(b)
	stream.Close()
	session.Close()

	// request := []byte("Hello, server! ")

	// for i := 0; i < 1024*2; i++ {
	// 	request = append(request, " Hello, server!"...)
	// }

	// chunks := splitMessage(request)

	// for _, chunk := range chunks {
	// 	// Send the request to the server
	// 	_, err = conn.Write(chunk)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

	// // Receive a response from the server
	// buffer := make([]byte, 1024)
	// n, _, err := conn.ReadFrom(buffer)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Received response:", string(buffer[:n]))
}

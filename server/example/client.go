package main

import (
	"fmt"
	"net"
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
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:12345")
	if err != nil {
		panic(err)
	}

	// Create a UDP connection to the server
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	request := []byte("Hello, server! ")

	for i := 0; i < 1024*2; i++ {
		request = append(request, " Hello, server!"...)
	}

	chunks := splitMessage(request)

	for _, chunk := range chunks {
		// Send the request to the server
		_, err = conn.Write(chunk)
		if err != nil {
			panic(err)
		}
	}

	// Receive a response from the server
	buffer := make([]byte, 1024)
	n, _, err := conn.ReadFrom(buffer)
	if err != nil {
		panic(err)
	}
	fmt.Println("Received response:", string(buffer[:n]))
}

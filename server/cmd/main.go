package main

import (
	"flag"

	"github.com/isaiahwong/cz4013/cmd/flight_client"
	"github.com/isaiahwong/cz4013/cmd/server"
	"github.com/isaiahwong/cz4013/common"
	"github.com/manifoldco/promptui"
)

// prompt provides an interactive prompt to configure the server and client
func prompt() {
	s := "Load Server"
	c := "Load Client"

	sp := promptui.Select{
		Label: "Select option",
		Items: []string{
			s,
			c,
		},
	}
	_, in, err := sp.Run()
	if common.HandleInterrupt(err) != nil {
		panic(err)
	}

	switch in {
	case s:
		server.Start()
	default:
		flight_client.Start()
	}
}

// runClient starts the client
func runClient() {
	flight_client.Start()
}

// runServer starts the server with the specified parameters
func runServer(deadline int, semantics int, port string, lossRate int) {
	s := server.New(semantics, deadline, lossRate, port)
	s.Serve()
}

// Entry point to application. Parses command line arguments and starts server or client
// The default mode if no flags are parsed are to run the server in AtMostOnce
func main() {
	var interactive bool
	var semantics int
	var deadline int
	var lossRate int
	var port string
	var client bool

	// Setup command line arguments
	flag.BoolVar(&interactive, "i", false, "Enables interactive mode. Other options will be ignored when interactive mode is enabled.")
	flag.BoolVar(&client, "c", false, "Run client")
	flag.IntVar(&deadline, "deadline", 5, "Deadline of a request response in seconds")
	flag.IntVar(&semantics, "semantic", 1, "[Server] Semantics of server. 0: AtLeastOnce, 1: AtMostOnce")
	flag.StringVar(&port, "port", "8080", "[Server] Server's port")
	flag.IntVar(&lossRate, "loss", 0, "[Server] Server's loss rate")

	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	// Start application in interactive mode if specified from prompt
	if interactive {
		prompt()
		return
	}

	// Starts application in client mode if specified from prompt
	if client {
		runClient()
		return
	}

	// Default runs to server
	runServer(deadline, semantics, port, lossRate)
}

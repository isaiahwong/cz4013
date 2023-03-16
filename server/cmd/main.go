package main

import (
	"flag"

	"github.com/isaiahwong/cz4013/cmd/flight_client"
	"github.com/isaiahwong/cz4013/cmd/server"
	"github.com/isaiahwong/cz4013/common"
	"github.com/manifoldco/promptui"
)

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

func runClient() {
	flight_client.Start()
}

func runServer(deadline int, semantics int, port string, lossRate int) {
	s := server.New(semantics, deadline, lossRate, port)
	s.Serve()
}

// Entry point to application. Parses command line arguments and starts server or client
func main() {
	var interactive bool
	var semantics int
	var deadline int
	var lossRate int
	var port string
	var client bool
	// var address string
	// var retries int

	// Setup command line arguments
	flag.BoolVar(&interactive, "i", false, "Enables interactive mode. Other options will be ignored when interactive mode is enabled.")
	flag.BoolVar(&client, "c", false, "Run client")
	flag.IntVar(&deadline, "deadline", 5, "Deadline of a request response in seconds")
	flag.IntVar(&semantics, "semantic", 1, "[Server] Semantics of server. 0: AtLeastOnce, 1: AtMostOnce")
	flag.StringVar(&port, "port", "8080", "[Server] Server's port")
	flag.IntVar(&lossRate, "loss", 0, "[Server] Server's loss rate")

	// flag.IntVar(&retries, "retries", 5, "[Client] Client retries when requests fails")
	// flag.StringVar(&address, "remote", "localhost:8080", "[Client] Remote address for client")

	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	if interactive {
		prompt()
		return
	}

	if client {
		runClient()
		return
	}

	runServer(deadline, semantics, port, lossRate)
}

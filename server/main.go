package main

import (
	"github.com/isaiahwong/cz4013/cmd/flight_client"
	"github.com/isaiahwong/cz4013/cmd/server"
	"github.com/isaiahwong/cz4013/common"
	"github.com/manifoldco/promptui"
)

func main() {
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

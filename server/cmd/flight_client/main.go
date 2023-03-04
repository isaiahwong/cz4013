package main

import (
	"time"

	"github.com/isaiahwong/cz4013/cmd/flight_client/app"
	"github.com/isaiahwong/cz4013/cmd/flight_client/client"
)

var c *client.Client
var a *app.App

func init() {
	c = client.New(
		client.WithAddr("localhost:8080"),
		client.WithDeadline(5*time.Second),
	)
	a = app.New(c)
}

func main() {
	if err := a.Start(); err != nil {
		panic(err)
	}
}

package flight_client

import (
	"time"

	"github.com/isaiahwong/cz4013/cmd/flight_client/app"
	"github.com/isaiahwong/cz4013/cmd/flight_client/client"
	"github.com/sirupsen/logrus"
)

var c *client.Client
var a *app.App

func init() {
	c = client.New(
		client.WithAddr("localhost:8080"),
		client.WithDeadline(5*time.Second),
		client.WithRetries(5),
		client.WithLogger(logrus.New()),
	)
	a = app.New(c)
}

func Start() {
	if err := a.Start(); err != nil {
		panic(err)
	}
}

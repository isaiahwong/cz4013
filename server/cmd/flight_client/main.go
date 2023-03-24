package flight_client

import (
	"strconv"
	"time"

	"github.com/isaiahwong/cz4013/cmd/flight_client/app"
	"github.com/isaiahwong/cz4013/cmd/flight_client/client"
	"github.com/isaiahwong/cz4013/common"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
)

var c *client.Client
var a *app.App

func prompt() *client.Client {
	loadDefault := "Load default config"
	customConfig := "Custom config"
	sp := promptui.Select{
		Label: "Select option",
		Items: []string{
			loadDefault,
			customConfig,
		},
	}

	_, input, err := sp.Run()
	if common.HandleInterrupt(err) != nil {
		panic(err)
	}

	if input == loadDefault {
		return client.New(
			client.WithAddr("localhost:8080"),
			client.WithDeadline(2*time.Second),
			client.WithRetries(5),
			client.WithLogger(logrus.New()),
		)
	}

	remoteAddrP := promptui.Prompt{
		Label: "Enter remote address <IP>:<PORT>",
	}
	remoteAddr, err := remoteAddrP.Run()
	if common.HandleInterrupt(err) != nil {
		panic(err)
	}

	timeoutP := promptui.Prompt{
		Label:    "Enter timeout in seconds",
		Validate: common.ValidateRange(1, 10),
	}
	timeout, err := timeoutP.Run()
	if common.HandleInterrupt(err) != nil {
		panic(err)
	}
	timeoutInt, _ := strconv.ParseInt(timeout, 10, 32)

	retriesP := promptui.Prompt{
		Label:    "Enter retries",
		Validate: common.ValidateRange(1, 10),
	}
	retry, err := retriesP.Run()
	if common.HandleInterrupt(err) != nil {
		panic(err)
	}
	retryInt, _ := strconv.ParseInt(retry, 10, 32)

	return client.New(
		client.WithAddr(remoteAddr),
		client.WithDeadline(time.Duration(timeoutInt)*time.Second),
		client.WithRetries(int(retryInt)),
		client.WithLogger(logrus.New()),
	)
}

func Start() {
	c = prompt()
	a = app.New(c)
	if err := a.Start(); err != nil {
		panic(err)
	}
}

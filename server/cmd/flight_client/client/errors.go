package client

import (
	"errors"
	"fmt"
)

func InvalidParams(msg string) error {
	return errors.New(fmt.Sprintf("Invalid params: %v", msg))
}

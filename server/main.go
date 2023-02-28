package main

import (
	"time"

	"github.com/isaiahwong/cz4013/protocol"
	"github.com/isaiahwong/cz4013/store"
)

func main() {
	db := store.New()
	s := protocol.New(
		protocol.WithDeadline(5*time.Second),
		protocol.WithDB(db),
	)
	s.Serve()
}

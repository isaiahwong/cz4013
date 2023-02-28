package main

import (
	"time"

	"github.com/isaiahwong/cz4013/protocol"
)

func main() {
	// p := Person{
	// 	Name:    "John",
	// 	Friends: []*Person{{Name: "Bob"}, {Name: "Alice"}},
	// }
	// b, err := encoding.Marshal(p)
	// if err != nil {
	// 	panic(err)
	// }

	// var s Person
	// err = encoding.Unmarshal(b, &s)
	// if err != nil {
	// 	panic(err)
	// }

	s := protocol.New(protocol.WithDeadline(5 * time.Second))
	s.Serve()

	// s, err := protocol.New(``
	// 	protocol.WithPort(":12345"),
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// s.Serve()
}

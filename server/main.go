package main

import (
	"fmt"

	"github.com/isaiahwong/cz4013/encoding"
)

type Person struct {
	Name    string
	Friends []*Person
}

func main() {
	p := Person{
		Name:    "John",
		Friends: []*Person{{Name: "Bob"}, {Name: "Alice"}},
	}
	b, err := encoding.Marshal(p)
	if err != nil {
		panic(err)
	}
	var s Person
	err = encoding.Unmarshal(b, &s)
	if err != nil {
		panic(err)
	}
	fmt.Println(s)

	// s, err := protocol.New(
	// 	protocol.WithPort(":12345"),
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// s.Serve()
}

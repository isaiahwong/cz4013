package main

import (
	"fmt"

	"github.com/isaiahwong/cz4013/encoding"
)

type Person struct {
	Name   string
	Friend *Person
	_      string
}

func main() {
	p := Person{Name: "John", Friend: &Person{Name: "John", Friend: nil}}
	b, err := encoding.Marshal(p)

	if err != nil {
		panic(err)
	}
	fmt.Println(b)

	// s, err := protocol.New(
	// 	protocol.WithPort(":12345"),
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// s.Serve()
}

package rpc

type Person struct {
	Name    string
	Friends []*Person
}

type Message struct {
	RPC  string
	Body []byte
}

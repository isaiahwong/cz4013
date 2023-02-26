package rpc

type Person struct {
	Name    string
	Friends []*Person
}

type Message struct {
	Sent int32
	RPC  string
	Body []byte
}

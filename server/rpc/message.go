package rpc

type Message struct {
	Sent int32
	RPC  string
	Body []byte
}

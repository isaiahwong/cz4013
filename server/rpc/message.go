package rpc

type Message struct {
	RPC   string
	Query map[string]string
	Body  []byte
}

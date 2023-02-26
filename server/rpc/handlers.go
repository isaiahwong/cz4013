package rpc

import "fmt"

type RPC struct {
}

func (r *RPC) HandleRequest(req *Message) *Message {
	fmt.Println(req)
	return nil
}

func New() *RPC {
	return &RPC{}
}

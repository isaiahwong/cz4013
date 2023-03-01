package rpc

type Error struct {
	Error string
	Body  string
}

type Message struct {
	RPC   string
	Query map[string]string
	Body  []byte
	Error *Error
}

func NewMessage(rpc string, body []byte) *Message {
	return &Message{
		RPC:  rpc,
		Body: body,
	}
}

func NewError(method string, err error, body string) *Message {
	return &Message{
		RPC: method,
		Error: &Error{
			Error: err.Error(),
			Body:  body,
		},
	}
}

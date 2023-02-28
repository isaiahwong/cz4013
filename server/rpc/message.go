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

func NewError(err error, body string) *Message {
	return &Message{
		Error: &Error{
			Error: err.Error(),
			Body:  body,
		},
	}
}

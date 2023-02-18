package protocol

import "github.com/sirupsen/logrus"

type options struct {
	logger *logrus.Logger
	port   string
}

// Option sets options for Server.
type Option func(*options)

// WithAddress returns an Option which sets the address the server will be listening to.
func WithPort(port string) Option {
	return func(o *options) {
		o.port = port
	}
}

// WithLogger returns an Option sets logger for gateway
func WithLogger(l *logrus.Logger) Option {
	return func(o *options) {
		o.logger = l
	}
}

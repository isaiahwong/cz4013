package client

import (
	"time"

	"github.com/sirupsen/logrus"
)

type options struct {
	logger   *logrus.Logger
	deadline time.Duration
	addr     string
}

// Option sets options for Server.
type Option func(*options)

// WithAddress returns an Option which sets the address the server will be listening to.
func WithAddr(a string) Option {
	return func(o *options) {
		o.addr = a
	}
}

func WithDeadline(deadline time.Duration) Option {
	return func(o *options) {
		o.deadline = deadline
	}
}

// WithLogger returns an Option sets logger for gateway
func WithLogger(l *logrus.Logger) Option {
	return func(o *options) {
		o.logger = l
	}
}

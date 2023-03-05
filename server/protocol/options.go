package protocol

import (
	"time"

	"github.com/isaiahwong/cz4013/rpc"
	"github.com/sirupsen/logrus"
)

type options struct {
	logger          *logrus.Logger
	port            string
	semantic        Semantics
	deadline        time.Duration
	flightRepo      *rpc.FlightRepo
	reservationRepo *rpc.ReservationRepo
	lossRate        int
}

// Option sets options for Server.
type Option func(*options)

// WithAddress returns an Option which sets the address the server will be listening to.
func WithPort(port string) Option {
	return func(o *options) {
		o.port = port
	}
}

func WithSemantic(semantic Semantics) Option {
	return func(o *options) {
		o.semantic = semantic
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

func WithFlightRepo(f *rpc.FlightRepo) Option {
	return func(o *options) {
		o.flightRepo = f
	}
}

func WithReservationRepo(r *rpc.ReservationRepo) Option {
	return func(o *options) {
		o.reservationRepo = r
	}
}

func WithLossRate(rate int) Option {
	return func(o *options) {
		o.lossRate = rate
	}
}

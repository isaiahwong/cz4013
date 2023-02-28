package rpc

type Flight struct {
	ID              int32
	Source          string
	Destination     string
	Timestamp       uint32
	Airfare         float32
	SeatAvailablity uint32
}

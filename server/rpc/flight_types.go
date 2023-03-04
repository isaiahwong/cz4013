package rpc

import "strconv"

type Flight struct {
	ID              int32
	Source          string
	Destination     string
	Airfare         float32
	SeatAvailablity int32
	Timestamp       uint32
}

type ReserveFlight struct {
	ID           string
	Flight       *Flight
	SeatReserved int32
	CheckIn      bool
}

func (f *Flight) Parse(data []string) error {
	id, err := strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return err
	}

	timestamp, err := strconv.ParseInt(data[3], 10, 64)
	if err != nil {
		return err
	}

	airfare, err := strconv.ParseFloat(data[4], 32)
	if err != nil {
		return err
	}

	seatAvail, err := strconv.ParseInt(data[5], 10, 64)
	if err != nil {
		return err
	}

	f.ID = int32(id)
	f.Source = data[1]
	f.Destination = data[2]
	f.Timestamp = uint32(timestamp)
	f.Airfare = float32(airfare)
	f.SeatAvailablity = int32(seatAvail)
	return nil
}

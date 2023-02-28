package rpc

import "strconv"

type Flight struct {
	ID              int32
	Source          string
	Destination     string
	Timestamp       uint32
	Airfare         float32
	SeatAvailablity uint32
}

func (f *Flight) Parse(data []string) error {
	id, err := strconv.Atoi(data[0])
	if err != nil {
		return err
	}

	timestamp, err := strconv.Atoi(data[3])
	if err != nil {
		return err
	}

	airfare, err := strconv.ParseFloat(data[4], 32)
	if err != nil {
		return err
	}

	seatAvail, err := strconv.Atoi(data[5])
	if err != nil {
		return err
	}

	f.ID = int32(id)
	f.Source = data[1]
	f.Destination = data[2]
	f.Timestamp = uint32(timestamp)
	f.Airfare = float32(airfare)
	f.SeatAvailablity = uint32(seatAvail)
	return nil
}

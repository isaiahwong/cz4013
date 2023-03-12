package rpc

import (
	"fmt"
	"strconv"
	"time"

	"github.com/isaiahwong/cz4013/common"
)

type Flight struct {
	ID              int32
	Source          string
	Destination     string
	Airfare         float32
	SeatAvailablity int32
	Timestamp       uint32
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

// String returns a string representation of the flight
func (f *Flight) String() string {
	t := time.Unix(int64(f.Timestamp), 0)
	localTime := t.Local()

	return fmt.Sprintf("%v%v%v%v%v%v",
		common.TitleValueLine("ID", f.ID, 1),
		common.TitleValueLine("Source", f.Source, 1),
		common.TitleValueLine("Destination", f.Destination, 1),
		common.TitleValueLine("Airfare", f.Airfare, 1),
		common.TitleValueLine("SeatAvailablity", f.SeatAvailablity, 1),
		common.TitleValueLine("Timestamp", localTime.Format("2006-01-02 15:04:05"), 1),
	)
}

type Food struct {
	ID   int64
	Name string
}

func (r *Food) String() string {
	return fmt.Sprintf("%v%v",
		common.TitleValueLine("ID", r.ID, 1),
		common.TitleValueLine("Name", r.Name, 1),
	)
}

func GetFood() map[int64]*Food {
	foodMap := map[int64]*Food{}
	meals := []*Food{
		{
			0, "Steak",
		},
		{
			1, "Pork Chop",
		},
		{
			2, "Wine",
		},
		{
			3, "Coke",
		},
	}
	for _, f := range meals {

		foodMap[f.ID] = f
	}
	return foodMap
}

type ReserveFlight struct {
	ID           string
	Flight       *Flight
	SeatReserved int32
	CheckIn      bool
	Cancelled    bool
	Meals        []*Food
}

func (r *ReserveFlight) String() string {
	return fmt.Sprintf("%v%v%v%v",
		common.TitleValueLine("ID", r.ID, 1),
		common.TitleValueLine("SeatReserved", r.SeatReserved, 1),
		common.TitleValueLine("Cancelled", r.Cancelled, 1),
		common.TitleValueLine("CheckIn", r.CheckIn, 1),
	)
}

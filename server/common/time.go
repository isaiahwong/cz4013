package common

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

func StrToUnixTime(timestamp string) (*time.Time, error) {
	// Check the number of digits in the timestamp value
	if len(timestamp) != 13 && len(timestamp) != 16 {
		return nil, errors.New("Invalid timestamp value")
	}

	// Determine the timestamp unit
	unit := time.Millisecond
	if len(timestamp) == 16 {
		unit = time.Microsecond
	}

	// Parse the timestamp value
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing timestamp value: %v", err)
	}

	// Convert the timestamp to a time.Time value
	t := time.Unix(0, ts*int64(unit))

	return &t, nil
}

package common

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
)

func CreateRegexp(s string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`(?i)\b\w*%v\w*\b`, s))
}

func TitleValueLine(title string, value any, numTabs int) string {
	tabs := strings.Repeat("\t", numTabs)
	return fmt.Sprintf("%v:%v%v\n", title, tabs, value)
}

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

func HandleInterrupt(err error) error {
	if err == promptui.ErrInterrupt {
		fmt.Println("Goodbye")
		os.Exit(0)
	}
	return err
}

func ValidateInt(input string) error {
	_, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return errors.New("Invalid input")
	}
	return nil
}

func NewLogger() *logrus.Logger {
	log := logrus.New()

	logrus.Trace()

	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(logrus.InfoLevel)
	// log.SetReportCaller(true)

	return log
}

func ValidateRange(l, r int64) func(input string) error {
	return func(input string) error {
		in, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return errors.New("Invalid input")
		}

		if in < l || in > r {
			return errors.New("Input out of range")
		}
		return nil
	}
}

func PrintTitle(title string) {
	fmt.Println("========================================")
	fmt.Println(title)
	fmt.Println("========================================")
}

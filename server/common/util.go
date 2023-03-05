package common

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

func CreateRegexp(s string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`(?i)\b\w*%v\w*\b`, s))
}

func TitleValueLine(title string, value any, numTabs int) string {
	tabs := strings.Repeat("\t", numTabs)
	return fmt.Sprintf("%v:%v%v\n", title, tabs, value)
}

func HandleInterrupt(err error) error {
	if err == promptui.ErrInterrupt {
		fmt.Println("Goodbye")
		os.Exit(0)
	}
	return err
}

var ValidateInt = func(input string) error {
	_, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return errors.New("Invalid input")
	}
	return nil
}

var ValidateRange = func(l, r int64) func(input string) error {
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

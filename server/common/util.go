package common

import (
	"fmt"
	"regexp"
)

func CreateRegexp(s string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`(?i)\b\w*%v\w*\b`, s))
}

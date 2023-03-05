package common

import (
	"fmt"
	"regexp"
	"strings"
)

func CreateRegexp(s string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`(?i)\b\w*%v\w*\b`, s))
}

func TitleValueLine(title string, value any, numTabs int) string {
	tabs := strings.Repeat("\t", numTabs)
	return fmt.Sprintf("%v:%v%v\n", title, tabs, value)
}

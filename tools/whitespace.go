package tools

import (
	"regexp"
	"strings"
)

var Whitespace = regexp.MustCompile("[ \\t\\r\\n]+")

func TrimWhitespace(s string) string {
	return strings.Trim(s, " \t\r\n")
}

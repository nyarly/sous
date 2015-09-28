package tools

import (
	"log"
	"os"
	"strings"
)

func TrimWhitespace(s string) string {
	return strings.Trim(s, " \t\r\n")
}

func Dief(format string, a ...interface{}) {
	Logf(format, a...)
	os.Exit(1)
}

func Logf(format string, a ...interface{}) {
	log.Printf(format, a...)
}

func ExitSuccessf(format string, a ...interface{}) {
	if len(format) != 0 {
		Logf(format, a...)
	}
	os.Exit(0)
}

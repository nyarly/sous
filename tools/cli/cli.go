package cli

import (
	"log"
	"os"
)

var _log *log.Logger

func init() {
	flags := 0
	if os.Getenv("TEAMCITY_VERSION") != "" {
		flags = log.Ldate | log.Ltime | log.Lmicroseconds
	}
	_log = log.New(os.Stderr, "", flags)
}

// Logf prints a formatted message to stderr
func Logf(format string, a ...interface{}) {
	_log.Printf(format, a...)
}

// Fatalf prints a formatted message to stderr and exits with exit code 1
func Fatalf(format string, a ...interface{}) {
	Logf(format, a...)
	os.Exit(1)
}

// Successf prints a formatted message to stderr and exits with exit code 0
func Successf(format string, a ...interface{}) {
	Logf(format, a...)
	Success()
}

func Success() {
	os.Exit(0)
}

package cli

import (
	"fmt"
	"log"
	"os"
)

var _log *log.Logger

func init() {
	flagss := 0
	// On TeamCity timestamp the logs
	if os.Getenv("TEAMCITY_VERSION") != "" {
		flagss = log.Ldate | log.Ltime | log.Lmicroseconds
		_log = log.New(os.Stderr, "", flagss)
	}
	if os.Getenv("DEBUG") == "YES" {
		flagss = log.LstdFlags | log.Lshortfile
		_log = log.New(os.Stderr, "", flagss)
	}
	_log = log.New(os.Stderr, "", flagss)
}

// Logf prints a formatted message to stderr
func Logf(format string, a ...interface{}) {
	_log.Print(ApplyStyles(fmt.Sprintf(format, a...)))
}

// Outf prints a formatted message to stdout
func Outf(format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf(format, a...))
}

// Fatalf prints a formatted message to stderr and exits with exit code 1
func Fatalf(format string, a ...interface{}) {
	Logf(format, a...)
	Fatal()
}

func Fatal() {
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

func LogBulletList(bullet string, list []string) {
	for _, item := range list {
		Logf("%s %s", bullet, item)
	}
}

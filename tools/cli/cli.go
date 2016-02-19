package cli

import (
	"fmt"
	"log"
	"os"
)

var _log *log.Logger

var beVerbose = false

func init() {
	flagss := 0
	// On TeamCity timestamp the logs
	if os.Getenv("TEAMCITY_VERSION") != "" {
		flagss |= log.Ldate | log.Ltime | log.Lmicroseconds
	}
	if os.Getenv("DEBUG") == "YES" {
		flagss |= log.LstdFlags | log.Lshortfile
	}
	_log = log.New(os.Stderr, "", flagss)
}

func BeVerbose() {
	beVerbose = true
}

// Logf prints a formatted message to stderr
func Logf(format string, a ...interface{}) {
	_log.Print(ApplyStyles(fmt.Sprintf(format, a...)))
}

// Verbosef calls Logf if verbose mode is on.
func Verbosef(format string, a ...interface{}) {
	if beVerbose {
		Logf(format, a...)
	}
}

// Curtf calls Logf if verbose mode is off. Often paired with a call to Verbosef.
func Curtf(format string, a ...interface{}) {
	if !beVerbose {
		Logf(format, a...)
	}
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

func Fatal(a ...interface{}) {
	if len(a) != 0 {
		fmt.Println(a...)
	}
	Exit(1)
}

func Exit(code int) {
	Cleanup()
	os.Exit(code)
}

// Successf prints a formatted message to stderr and exits with exit code 0
func Successf(format string, a ...interface{}) {
	Logf(format, a...)
	Success()
}

func Success() {
	Exit(0)
}

func LogBulletList(bullet string, list []string) {
	for _, item := range list {
		Logf("%s %s", bullet, item)
	}
}

func Warn(format string, a ...interface{}) {
	format = fmt.Sprintf("**sous> WARNING: %s **", format)
	Logf(format, a...)
}

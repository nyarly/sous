package main

import (
	"os"
	"time"

	"github.com/opentable/sous/config"
	"github.com/opentable/sous/tools/cli"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	loadCommands()
	command := os.Args[1]
	c, ok := Sous.Commands[command]
	if !ok {
		cli.Fatalf("Command %s not recognised; try `sous help`", command)
	}
	if command != "config" {
		updateHourly()
	}
	// It is the responsibility of the command to exit with an appropriate
	// error code...
	c.Func(buildPacks, os.Args[2:])
	// If it does not, we assume it failed...
	cli.Fatalf("Command did not complete correctly")
}

func usage() {
	cli.Fatalf("usage: sous COMMAND; try `sous help`")
}

func updateHourly() {
	key := "last-update-check"
	props := config.Properties()
	d, err := time.Parse(time.RFC3339, props[key])
	if err != nil || d.Sub(time.Now()) > time.Hour {
		checkForUpdates()
	}
}

func checkForUpdates() {
	cli.Logf("Checking for updates...")
	if err := config.Update(); err != nil {
		cli.Logf("Unable to check: %s", err)
	}
}

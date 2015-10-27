package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opentable/sous/config"
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
)

func main() {
	trapSignals()
	defer cli.Cleanup()
	if len(os.Args) < 2 {
		usage()
	}
	command := os.Args[1]
	if command != "config" {
		updateHourly()
	}
	cfg := config.Load()
	sous := core.NewSous(Version, Revision, OS, Arch, loadCommands(), BuildPacks(cfg))
	c, ok := sous.Commands[command]
	if !ok {
		cli.Fatalf("Command %s not recognised; try `sous help`", command)
	}
	// It is the responsibility of the command to exit with an appropriate
	// error code...
	c.Func(sous, os.Args[2:])
	// If it does not, we assume it failed...
	cli.Fatalf("Command did not complete correctly")
}

func usage() {
	cli.Fatalf("usage: sous <command>; try `sous help`")
}

// trapSignals traps both SIGINT and SIGTERM and defers to cli.Exit
// to do a graceful exit.
func trapSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c
		cli.Exit(128 + int(s.(syscall.Signal)))
	}()
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

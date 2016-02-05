package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/deploy"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/file"
)

//go:generate ./scripts/generate-resources core/resources

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	sousFlags, args := parseFlags(os.Args[2:])
	command := os.Args[1]
	var state *deploy.State
	var sous *core.Sous
	if command != "config" && command != "update" {
		updateHourly()
		file.ReadJSON(&state, "~/.sous/config")
		trapSignals()
		defer cli.Cleanup()
		sous = core.NewSous(Version, Revision, OS, Arch, loadCommands(), BuildPacks(&state.Config), sousFlags, state)
	} else {
		sous = core.NewSous(Version, Revision, OS, Arch, loadCommands(), nil, sousFlags, nil)
	}
	c, ok := sous.Commands[command]
	if !ok {
		cli.Fatalf("Command %s not recognised; try `sous help`", command)
	}
	// It is the responsibility of the command to exit with an appropriate
	// error code...
	c.Func(sous, args)
	// If it does not, we assume it failed...
	cli.Fatalf("Command did not complete correctly")
}

func parseFlags(args []string) (*core.SousFlags, []string) {
	flagSet := flag.NewFlagSet("sous", flag.ExitOnError)
	rebuild := flagSet.Bool("rebuild", false, "force a rebuild")
	rebuildAll := flagSet.Bool("rebuild-all", false, "force a rebuild of this target plus all dependencies")
	verbose := flagSet.Bool("v", false, "be verbose")
	err := flagSet.Parse(args)
	if err != nil {
		cli.Fatalf("%s", err)
	}
	if *verbose {
		cli.BeVerbose()
	}
	return &core.SousFlags{
		ForceRebuild:    *rebuild,
		ForceRebuildAll: *rebuildAll,
	}, flagSet.Args()
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
	props := deploy.Properties()
	d, err := time.Parse(time.RFC3339, props[key])
	if err != nil || d.Sub(time.Now()) > time.Hour {
		checkForUpdates()
	}
}

func checkForUpdates() {
	cli.Logf("Checking for updates...")
	if err := deploy.Update(); err != nil {
		cli.Logf("Unable to check: %s", err)
	}
}

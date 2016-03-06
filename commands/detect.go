package commands

import (
	"fmt"
	"os"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
)

func DetectHelp() string {
	return `detect detects available targets for your project, and tells you how to enable those targets.

A note on exit codes: Detect returns a success exit code if any project is detected, whether or not it supports any targets.`
}

func Detect(sous *core.Sous, args []string) {
	c := core.GetContext()
	var pack *core.Buildpack
	// First determine the pack if possible
	if len(args) == 0 {
		pack = c.DetectProjectType(sous.State.Buildpacks)
		if pack == nil {
			fmt.Println("no sous-compatible project detected")
			os.Exit(1)
		}
	} else if len(args) == 1 {
		var ok bool
		pack, ok = sous.State.Buildpacks.Get(args[0])
		if !ok {
			cli.Fatalf("buildpack %q does not exist", args[0])
		}
		_, err := pack.Detect(c.WorkDir)
		if err != nil {
			cli.Fatal(err)
		}
	}

	// Next determine stack version compat
	runnable, err := pack.BindStackVersion(c.WorkDir)
	if err != nil {
		cli.Logf("Detected a %s %s project.", pack.Name, pack.DetectedStackVersionRange)
		cli.Fatal(err)
	}

	actualVersion, err := runnable.StackVersion.Version()
	if err != nil {
		cli.Fatal(err)
	}

	cli.Outf("Detected a %s %s (%s) project.",
		pack.Name, runnable.DetectedStackVersionRange, actualVersion.String())

	cli.Outf("Build Version: %s", c.BuildVersion)

	cli.Success()
}

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
	pack := c.DetectProjectType(sous.State.Buildpacks)
	if pack == nil {
		fmt.Println("no sous-compatible project detected")
		os.Exit(1)
	}
	cli.Outf("Detected a %s %s project.", pack.Detect(c.WorkDir))
	cli.Outf("Build Version: %s", c.BuildVersion)
	cli.Success()
}

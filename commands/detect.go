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
	pack := core.DetectProjectType(sous.Packs)
	if pack == nil {
		fmt.Println("no sous-compatible project detected")
		os.Exit(1)
	}
	if fatal := core.CheckForProblems(pack); fatal {
		cli.Fatalf("Detected a %s project with fatal errors.", pack)
	}
	c := core.GetContext("app")
	desc := pack.AppDesc()
	cli.Outf("Detected a %s; which supports the following targets...", desc)
	for _, target := range pack.Targets() {
		if err := target.Check(); err != nil {
			cli.Outf("\t%s \t✘ %s", target, err)
			continue
		}
		cli.Outf("\t%s \t✔︎", target)
	}
	cli.Outf("Build Version: %s", c.BuildVersion)
	cli.Success()
}

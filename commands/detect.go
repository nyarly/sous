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
	problems := pack.Problems()
	if len(problems) != 0 {
		cli.Outf("Detected a %s project with some issues...", pack)
		cli.LogBulletList("-", problems)
		cli.Fatal()
	}
	desc := pack.AppDesc()
	cli.Outf("Detected %s; target support...", desc)
	for _, target := range pack.Targets() {
		if err := target.Check(); err != nil {
			cli.Outf("\t%s \t✘ %s", target, err)
			continue
		}
		cli.Outf("\t%s \t✔︎", target)
	}
	os.Exit(0)
}

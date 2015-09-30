package commands

import (
	"fmt"
	"os"

	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
)

func Detect(packs []*build.Pack, args []string) {
	pack := build.DetectProjectType(packs)
	if pack == nil {
		fmt.Println("no sous-compatible project detected")
		os.Exit(1)
	}
	incompatabilities := pack.CheckCompatibility()
	if len(incompatabilities) != 0 {
		cli.Outf("Detected a %s project\n", pack.Name)
		cli.Logf("You need to fix a few things before you can build this project..")
		for _, message := range incompatabilities {
			cli.Logf("\t%s", message)
		}
		cli.Fatal()
	}
	desc := pack.CompatibleProjectDesc()
	cli.Outf("Detected %s; target support...", desc)
	context := build.GetContext("detect")
	for name, feature := range pack.Features {
		if _, err := feature.Detect(context); err != nil {
			cli.Outf("\t%s \t✘ %s", name, err)
			continue
		}
		cli.Outf("\t%s \t✔︎", name)
	}
	os.Exit(0)
}

func DetectHelp() string {
	return `detect detects available actions for your project, and
tells you how to enable those features`
}

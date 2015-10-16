package commands

import (
	"fmt"
	"os"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
)

func Detect(sous *core.Sous, args []string) {
	pack := core.DetectProjectType(sous.Packs)
	if pack == nil {
		fmt.Println("no sous-compatible project detected")
		os.Exit(1)
	}
	incompatabilities := pack.CheckCompatibility()
	if len(incompatabilities) != 0 {
		cli.Outf("Detected a %s project with some issues...", pack.Name)
		cli.LogBulletList("-", incompatabilities)
		cli.Fatal()
	}
	desc := pack.CompatibleProjectDesc()
	cli.Outf("Detected %s; target support...", desc)
	context := core.GetContext("detect")
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

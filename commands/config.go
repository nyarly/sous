package commands

import (
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
)

func ConfigHelp() string {
	return "sous config gets and sets config properties for your sous installation"
}

func Config(sous *core.Sous, args []string) {
	if len(args) == 0 || len(args) > 2 {
		cli.Fatalf("usage: sous config <key> [<new-value>]")
	}
	if len(args) == 1 {
		if v, ok := core.Properties()[args[0]]; ok {
			cli.Outf(v)
			cli.Success()
		}
		cli.Fatalf("Key %s not found", args[0])
	}
	core.Set(args[0], args[1])
	cli.Logf("Successfully set %s to %s", args[0], args[1])
	cli.Success()
}

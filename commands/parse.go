package commands

import (
	"os"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/yaml"
)

func ParseStateHelp() string {
	return `sous parse-state parses a sous state directory hierarchy`
}

func ParseState(sous *core.Sous, args []string) {
	stateDir := getStateDir(args)
	state, err := core.Parse(stateDir)
	if err != nil {
		cli.Fatalf("%s", err)
	}
	merged, err := state.Merge()
	if err != nil {
		cli.Fatalf("%s", err)
	}
	out, err := yaml.Marshal(merged)
	if err != nil {
		cli.Fatalf("%s", err)
	}
	cli.Outf(string(out))
	cli.Success()
}

func getStateDir(args []string) string {
	if len(args) != 0 {
		return args[0]
	} else {
		d, err := os.Getwd()
		if err != nil {
			cli.Fatalf("%s", err)
		}
		return d
	}
}

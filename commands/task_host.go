package commands

import (
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
)

func TaskHost(sous *core.Sous, args []string) {
	cli.Outf("%s", core.DivineTaskHost())
	cli.Success()
}

func TaskHostHelp() string {
	return "task host used for contracts (typically your active docker host)"
}

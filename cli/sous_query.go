package cli

import (
	"flag"

	"github.com/opentable/sous/util/cmdr"
)

type SousQuery struct {
	Sous  *Sous
	flags struct {
		target              string
		rebuild, rebuildAll bool
	}
}

var QuerySubcommands = cmdr.Commands{}

func init() { TopLevelCommands["query"] = &SousQuery{} }

const sousQueryHelp = `
inquire about sous details
`

func (*SousQuery) Help() string { return sousBuildHelp }

func (sb *SousQuery) AddFlags(fs *flag.FlagSet) {
}

func (SousQuery) Subcommands() cmdr.Commands {
	return QuerySubcommands
}

func (sb *SousQuery) Execute(args []string) cmdr.Result {
	err := UsageErrorf("usage: sous query [options] command")
	err.Tip = "try `sous query help` for a list of commands"
	return err
}

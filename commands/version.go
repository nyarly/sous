package commands

import (
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
)

func Version(sous *core.Sous, args []string) {
	cli.Outf("Sous version %s %s/%s", sous.Version, sous.OS, sous.Arch)
	cli.Outf("Revision: %s", sous.Revision)
	cli.Success()
}

func VersionHelp() string {
	return "Sous version information"
}

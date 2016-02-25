package commands

import (
	"fmt"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/dir"
)

func BuildPathHelp() string {
	return `sous build-path prints the working directory for this project's build state`
}

func BuildPath(sous *core.Sous, args []string) {
	target := "app"
	if len(args) != 0 {
		target = args[0]
	}
	tc := sous.TargetContext(target)
	fmt.Println(dir.Resolve(dir.DirName(tc.BaseDir())))
	cli.Success()
}

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
	_, context := sous.AssembleTargetContext(target)
	fmt.Println(dir.Resolve(dir.Dir(context.BaseDir())))
	cli.Success()
}

package commands

import (
	"fmt"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/path"
)

func BuildPathHelp() string {
	return `sous build-path prints the working directory for this project's build state`
}

func BuildPath(sous *core.Sous, args []string) {
	_, context, _ := AssembleFeatureContext("build", sous.Packs)
	fmt.Println(path.Resolve(path.BaseDir(context.BaseDir())))
	cli.Success()
}

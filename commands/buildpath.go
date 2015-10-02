package commands

import (
	"fmt"

	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/path"
)

func BuildPathHelp() string {
	return `sous build-path prints the working directory for this project's build state`
}

func BuildPath(packs []*build.Pack, args []string) {
	_, context, _ := AssembleFeatureContext("build", packs)
	fmt.Println(path.Resolve(path.BaseDir(context.BaseDir())))
	cli.Success()
}

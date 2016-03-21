package commands

import (
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
)

func DockerfileHelp() string {
	return `sous dockerfile prints the current dockerfile for this project`
}

func Dockerfile(sous *core.Sous, args []string) {
	targetName := "app"
	if len(args) != 0 {
		targetName = args[0]
	}
	target, context := sous.AssembleTargetContext(targetName)
	cli.Outf(sous.Dockerfile(target, context).Render())
	cli.Success()
}

package commands

import (
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
)

func RunHelp() string {
	return `sous run your project (building first if necessary)`
}

func Run(sous *core.Sous, args []string) {
	targetName := "app"
	if len(args) != 0 {
		targetName = args[0]
	}
	core.RequireGit()
	core.RequireDocker()

	target, context := sous.AssembleTargetContext(targetName)

	sous.RunTarget(target, context)

	var dr *docker.Run
	runner, ok := target.(core.ContainerTarget)
	if !ok {
		cli.Fatalf("%s->%s does not support running", target.Pack(), target.Name())
	} else {

	}
	dr = runner.DockerRun(context)
	if code := dr.ExitCode(); code != 0 {
		cli.Fatalf("Run failed with exit code %d", code)
	}
	cli.Success()
}

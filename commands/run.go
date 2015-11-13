package commands

import (
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
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
	runner, ok := target.(core.ContainerTarget)
	if !ok {
		cli.Fatalf("Target %s does not support running.", target.Name())
	}

	rebuilt, _ := sous.RunTarget(target, context)
	dr, _ := sous.RunContainerTarget(runner, context, rebuilt)
	if exitCode := dr.ExitCode(); exitCode != 0 {
		cli.Logf("Docker container exited with code %d", exitCode)
		cli.Exit(exitCode)
	}
	cli.Success()
}

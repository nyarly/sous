package commands

import (
	"strconv"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/ports"
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
	if !sous.BuildIfNecessary(target, context) {
		cli.Logf("No relevant changes since last build, running %s", context.DockerTag())
	}
	var dr *docker.Run
	if runner, ok := target.(core.DockerRunner); ok {
		dr = runner.DockerRun(context)
	} else {
		dr = defaultDockerRun(context)
	}
	if code := dr.ExitCode(); code != 0 {
		cli.Fatalf("Run failed with exit code %d", code)
	}
	cli.Success()
}

func defaultDockerRun(context *core.Context) *docker.Run {
	dr := docker.NewRun(context.DockerTag())
	port0, err := ports.GetFreePort()
	if err != nil {
		cli.Fatalf("Unable to get free port: %s", err)
	}
	dr.AddEnv("PORT0", strconv.Itoa(port0))
	dr.AddEnv("TASK_HOST", core.DivineTaskHost())
	return dr
}

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
	targetName := "build"
	if len(args) != 0 {
		targetName = args[0]
	}
	core.RequireGit()
	core.RequireDocker()

	target, context, appInfo := sous.AssembleTargetContext(targetName)
	if !sous.BuildIfNecessary(target, context, appInfo) {
		cli.Logf("No changes since last build, running %s", context.DockerTag())
	}

	dr := docker.NewRun(context.DockerTag())
	port0, err := ports.GetFreePort()
	if err != nil {
		cli.Fatalf("Unable to get free port: %s", err)
	}
	dr.AddEnv("PORT0", strconv.Itoa(port0))
	dr.AddEnv("TASK_HOST", core.DivineTaskHost())
	if code := dr.ExitCode(); code != 0 {
		cli.Fatalf("Run failed with exit code %d", code)
	}
	cli.Success()
}

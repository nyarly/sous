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
	target := "build"
	if len(args) != 0 {
		target = args[0]
	}
	RequireGit()
	RequireDocker()

	feature, context, appInfo := AssembleFeatureContext(target, sous.Packs)
	if !BuildIfNecessary(feature, context, appInfo) {
		cli.Logf("No changes since last build, running %s", context.DockerTag())
	}

	dr := docker.NewRun(context.DockerTag())
	port0, err := ports.GetFreePort()
	if err != nil {
		cli.Fatalf("Unable to get free port: %s", err)
	}
	dr.AddEnv("PORT0", strconv.Itoa(port0))
	dr.AddEnv("TASK_HOST", divineTaskHost())
	if code := dr.ExitCode(); code != 0 {
		cli.Fatalf("Run failed with exit code %d", code)
	}
	cli.Success()
}

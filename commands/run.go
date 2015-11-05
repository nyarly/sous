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
	}
	dr = runner.DockerRun(context)
	container, err := dr.Start()
	if err != nil {
		cli.Fatalf("Unable to start container: %s", err)
	}
	cli.AddCleanupTask(func() error {
		if container.Exists() {
			if container.Running() {
				if err := container.Kill(); err != nil {
					return err
				}
			}
		} else {
			cli.Logf("Attempted to clean up container %s, but it doesn't exist", container)
		}
		return nil
	})
	if err := container.Wait(); err != nil {
		cli.Fatalf("Run failed: %s", err)
	}
	cli.Success()
}

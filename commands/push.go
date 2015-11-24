package commands

import (
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/git"
)

func PushHelp() string {
	return `sous push pushes your latest build to the docker registry`
}

func Push(sous *core.Sous, args []string) {
	target := "app"
	if len(args) != 0 {
		target = args[0]
	}
	core.RequireGit()
	core.RequireDocker()
	if err := git.AssertCleanWorkingTree(); err != nil {
		cli.Warn("Dirty working tree: %s", err)
	}

	_, context := sous.AssembleTargetContext(target)

	tag := context.DockerTag()
	if !docker.ImageExists(tag) {
		cli.Fatalf("No built image available; try building first")
	}
	docker.Push(tag)
	name := context.CanonicalPackageName()
	cli.Successf("Successfully pushed %s v%s as %s", name, context.BuildVersion, context.DockerTag())
}

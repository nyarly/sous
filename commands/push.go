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

	tc := sous.TargetContext(target)

	tag := tc.DockerTag()
	if !docker.ImageExists(tag) {
		cli.Fatalf("No built image available; try building first")
	}
	docker.Push(tag)
	name := tc.CanonicalPackageName()
	cli.Successf("Successfully pushed %s v%s as %s", name, tc.BuildVersion, tc.DockerTag())
}

package commands

import (
	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/git"
)

func PushHelp() string {
	return `sous push pushes your latest build to the docker registry`
}

func Push(packs []*build.Pack, args []string) {
	target := "build"
	if len(args) != 0 {
		target = args[0]
	}
	RequireGit()
	RequireDocker()
	if err := git.AssertCleanWorkingTree(); err != nil {
		cli.Logf("WARNING: Dirty working tree: %s", err)
	}

	_, context, appInfo := AssembleFeatureContext(target, packs)

	lastBuildTag := context.PrevDockerTag()
	if !docker.ImageExists(lastBuildTag) {
		cli.Fatalf("No built image available; try building first")

	}
	docker.Push(lastBuildTag)
	name := context.CanonicalPackageName()
	cli.Successf("Successfully pushed %s v%s as %s", name, appInfo.Version, context.DockerTag())
}

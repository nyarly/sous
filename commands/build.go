package commands

import (
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/git"
)

func BuildHelp() string {
	return `sous build detects your project type, and tries to find a matching
stack to build against. Right now it only supports NodeJS projects. It builds a
docker image, tagged and labelled correctly.

sous build does not have any options yet`
}

func Build(sous *core.Sous, args []string) {
	targetName := "app"
	if len(args) != 0 {
		targetName = args[0]
	}
	core.RequireGit()
	core.RequireDocker()
	if err := git.AssertCleanWorkingTree(); err != nil {
		cli.Warn("Dirty working tree: %s", err)
	}

	tc := sous.TargetContext(targetName)

	built, _ := sous.RunTarget(tc)

	if !built {
		cli.Successf("Already built: %s", tc.DockerTag())
	}

	name := tc.CanonicalPackageName()
	cli.Successf("Successfully built %s v%s as %s", name, tc.BuildVersion, tc.DockerTag())
}

package commands

import (
	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/git"
)

func BuildHelp() string {
	return `sous build detects your project type, and tries to find a matching
OpenTable supported stack to build against. Right now it only supports NodeJS
projects. It builds a docker image, tagged and labelled correctly.

sous build does not have any options yet`
}

func Build(packs []*build.Pack, args []string) {
	RequireGit()
	RequireDocker()
	git.RequireCleanWorkingTree()

	feature, context, appInfo := AssembleFeatureContext("build", packs)
	if !BuildIfNecessary(feature, context, appInfo) {
		cli.Successf("Already built: %s", context.DockerTag())
	}
	name := context.CanonicalPackageName()
	cli.Successf("Successfully built %s v%s as %s", name, appInfo.Version, context.DockerTag())
}

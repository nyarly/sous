package commands

import (
	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
)

func TestHelp() string {
	return `sous build detects your project type, and tries to find a matching
OpenTable supported stack to build against. Right now it only supports NodeJS
projects. It builds a docker image, tagged and labelled correctly.

sous build does not have any options yet`
}

func Test(packs []*build.Pack, args []string) {
	RequireGit()
	RequireDocker()

	feature, context, appInfo := AssembleFeatureContext("test", packs)
	if !BuildIfNecessary(feature, context, appInfo) {
		cli.Logf("No changes since last build, running %s", context.DockerTag())
	}

	testRunExitCode := docker.Run(context.DockerTag())

	if testRunExitCode == 0 {
		name, version := context.CanonicalPackageName(), appInfo.Version
		cli.Successf("Tests passed %s v%s as %s", name, version, context.DockerTag())
	}
	cli.Fatalf("Test(s) failed.")
}

package commands

import (
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
)

func TestHelp() string {
	return `sous build detects your project type, and tries to find a matching
OpenTable supported stack to build against. Right now it only supports NodeJS
projects. It builds a docker image, tagged and labelled correctly.

sous build does not have any options yet`
}

func Test(sous *core.Sous, args []string) {
	core.RequireGit()
	core.RequireDocker()

	feature, context, appInfo := sous.AssembleFeatureContext("test")
	if !sous.BuildIfNecessary(feature, context, appInfo) {
		cli.Logf("No changes since last build, running %s", context.DockerTag())
	}

	testRunExitCode := docker.NewRun(context.DockerTag()).ExitCode()

	if testRunExitCode == 0 {
		name, version := context.CanonicalPackageName(), appInfo.Version
		cli.Successf("Tests passed %s v%s as %s", name, version, context.DockerTag())
	}
	cli.Fatalf("Test(s) failed.")
}

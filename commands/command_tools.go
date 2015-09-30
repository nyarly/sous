package commands

import (
	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/git"
	"github.com/opentable/sous/tools/version"
)

func AssembleFeatureContext(name string, packs []*build.Pack) (*build.Feature, *build.Context, *build.AppInfo) {
	pack := build.DetectProjectType(packs)
	if pack == nil {
		cli.Fatalf("no buildable project detected")
	}
	buildFeature, ok := pack.Features["build"]
	if !ok {
		cli.Fatalf("The %s build pack does not support build", pack.Name)
	}
	context := build.GetContext("build")
	appInfo, err := buildFeature.Detect(context)
	if err != nil {
		cli.Fatalf("unable to %s %s project: %s", "build", pack.Name, err)
	}
	context.AppVersion = appInfo.Version
	return buildFeature, context, appInfo
}

func RequireDocker() {
	docker.RequireVersion(version.Range(">=1.8.2"))
	docker.RequireDaemon()
}

func RequireGit() {
	git.RequireVersion(version.Range(">=2.0.0"))
	git.RequireRepo()
}

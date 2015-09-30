package commands

import (
	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/git"
	"github.com/opentable/sous/tools/version"
)

func TestHelp() string {
	return `sous build detects your project type, and tries to find a matching
OpenTable supported stack to build against. Right now it only supports NodeJS
projects. It builds a docker image, tagged and labelled correctly.

sous build does not have any options yet`
}

func Test(packs []*build.Pack, args []string) {

	git.RequireVersion(version.Range(">=2.0.0"))
	git.RequireRepo()
	docker.RequireVersion(version.Range(">=1.8.2"))
	docker.RequireDaemon()

	context := build.GetContext("test")
	pack := build.DetectProjectType(packs)
	if pack == nil {
		cli.Fatalf("no testable project detected")
	}
	buildFeature, ok := pack.Features["test"]
	if !ok {
		cli.Fatalf("The %s build pack does not support test", pack.Name)
	}
	appInfo, err := buildFeature.Detect(context)
	context.AppVersion = appInfo.Version
	if err != nil {
		cli.Fatalf("unable to test %s project: %s", pack.Name, err)
	}
	df := buildFeature.MakeDockerfile(appInfo)
	addMetadata(df, context)

	tag := context.PrevDockerTag()
	if context.NeedsBuild() {
		tag = context.NextDockerTag()
		docker.Build(context.BaseDir(), tag)
		context.Commit()
	}

	docker.Run(tag)

	cli.Successf("Successfully built %s v%s as %s",
		context.CanonicalPackageName(), appInfo.Version, tag)
}

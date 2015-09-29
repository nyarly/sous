package commands

import (
	"github.com/opentable/sous/build"
	. "github.com/opentable/sous/tools"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/file"
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

	context := build.GetContext()
	pack := build.DetectProjectType(packs)
	if pack == nil {
		Dief("no testable project detected")
	}
	buildFeature, ok := pack.Features["test"]
	if !ok {
		Dief("The %s build pack does not support test", pack.Name)
	}
	appInfo, err := buildFeature.Detect(context)
	if err != nil {
		Dief("unable to test %s project: %s", pack.Name, err)
	}
	df := buildFeature.MakeDockerfile(appInfo)
	addMetadata(df, context)
	file.Write(df.Render(), "Dockerfile")

	tag := dockerTag(context, appInfo, "-tests")

	docker.Build(tag)

	docker.Run(tag)

	ExitSuccessf("Successfully built %s v%s as %s",
		context.CanonicalPackageName(), appInfo.Version, tag)
}

package commands

import (
	"fmt"
	"os"

	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/git"
)

func BuildHelp() string {
	return `sous build detects your project type, and tries to find a matching
OpenTable supported stack to build against. Right now it only supports NodeJS
projects. It builds a docker image, tagged and labelled correctly.

sous build does not have any options yet`
}

func Build(packs []*build.Pack, args []string) {
	git.RequireCleanWorkingTree()
	RequireGit()
	RequireDocker()

	buildFeature, context, appInfo := AssembleFeatureContext("build", packs)
	if !context.NeedsBuild() {
		fmt.Printf("Already built: %s", context.PrevDockerTag())
		os.Exit(0)
	}

	df := buildFeature.MakeDockerfile(appInfo)
	addMetadata(df, context)
	context.SaveFile(df.Render(), "Dockerfile")

	tag := context.NextDockerTag()

	docker.Build(context.BaseDir(), tag)

	context.Commit()

	cli.Successf("Successfully built %s v%s as %s",
		context.CanonicalPackageName(), appInfo.Version, tag)
}

func addMetadata(d *docker.Dockerfile, c *build.Context) {
	d.Maintainer = c.User
	prefix := "com.opentable.build"
	d.AddLabel(prefix+".builder.app", "sous")
	d.AddLabel(prefix+".builder.host", c.Host)
	d.AddLabel(prefix+".builder.fullhost", c.FullHost)
	d.AddLabel(prefix+".builder.user", c.User)
	d.AddLabel(prefix+".source.git.repo", c.Git.CanonicalName())
	d.AddLabel(prefix+".source.git.commit-sha", c.Git.CommitSHA)
}

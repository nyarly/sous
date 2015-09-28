package commands

import (
	"fmt"

	"github.com/opentable/sous/build"
	. "github.com/opentable/sous/tools"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/file"
)

func BuildHelp() string {
	return `sous build detects your project type, and tries to find a matching
OpenTable supported stack to build against. Right now it only supports NodeJS
projects. It builds a docker image, tagged and labelled correctly.

sous build does not have any options yet`
}

func Build(packs []*build.Pack, args []string) {
	context := build.GetContext()
	pack := build.DetectProjectType(packs)
	if pack == nil {
		Dief("no buildable project detected")
	}
	buildFeature, ok := pack.Features["build"]
	if !ok {
		Dief("The %s build pack does not support build", pack.Name)
	}
	appInfo, err := buildFeature.Detect(context)
	if err != nil {
		Dief("unable to build %s project: %s", pack.Name, err)
	}
	df := buildFeature.MakeDockerfile(appInfo)
	addMetadata(df, context)
	file.Write(df.Render(), "Dockerfile")

	tag := dockerTag(context, appInfo)

	docker.Build(tag)

	ExitSuccessf("Successfully built %s v%s as %s",
		context.CanonicalPackageName(), appInfo.Version, tag)
}

func dockerTag(c *build.Context, a *build.AppInfo) string {
	// e.g. on TeamCity:
	//   docker.otenv.com/widget-factory:v0.12.1-ci-912eeeab-1
	if c.IsCI() {
		return fmt.Sprintf("%s/%s:v%s-ci-%s-%d",
			c.DockerRegistry,
			c.CanonicalPackageName(),
			a.Version,
			c.Git.CommitSHA[0:8],
			c.BuildNumber)
	}
	// e.g. on local dev machine:
	//   docker.otenv.com/widget-factory:username@host-v0.12.1-912eeeab-1
	return fmt.Sprintf("%s/%s/%s:v%s-%s-%s-%d",
		c.DockerRegistry,
		c.User,
		c.CanonicalPackageName(),
		a.Version,
		c.Git.CommitSHA[0:8],
		c.Host,
		c.BuildNumber)
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

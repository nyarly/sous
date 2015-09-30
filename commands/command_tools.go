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
	buildFeature, ok := pack.Features[name]
	if !ok {
		cli.Fatalf("The %s build pack does not support %s", pack.Name, name)
	}
	context := build.GetContext(name)
	appInfo, err := buildFeature.Detect(context)
	if err != nil {
		cli.Fatalf("unable to %s %s project: %s", name, pack.Name, err)
	}
	context.AppVersion = appInfo.Version
	return buildFeature, context, appInfo
}

func BuildIfNecessary(feature *build.Feature, context *build.Context, appInfo *build.AppInfo) bool {
	if !context.NeedsBuild() {
		context.BuildState.CurrentCommit().BuildNumber--
		return false
	}

	context.IncrementBuildNumber()

	df := feature.MakeDockerfile(appInfo)
	AddMetadata(df, context)
	context.SaveFile(df.Render(), "Dockerfile")
	docker.Build(context.BaseDir(), context.DockerTag())
	context.Commit()
	return true
}

func RequireDocker() {
	docker.RequireVersion(version.Range(">=1.8.2"))
	docker.RequireDaemon()
}

func RequireGit() {
	git.RequireVersion(version.Range(">=2.0.0"))
	git.RequireRepo()
}

func AddMetadata(d *docker.Dockerfile, c *build.Context) {
	d.Maintainer = c.User
	prefix := "com.opentable.build"
	d.AddLabel(prefix+".builder.app", "sous")
	d.AddLabel(prefix+".builder.host", c.Host)
	d.AddLabel(prefix+".builder.fullhost", c.FullHost)
	d.AddLabel(prefix+".builder.user", c.User)
	d.AddLabel(prefix+".source.git.repo", c.Git.CanonicalName())
	d.AddLabel(prefix+".source.git.commit-sha", c.Git.CommitSHA)
}

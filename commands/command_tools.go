package commands

import (
	"os"

	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/git"
	"github.com/opentable/sous/tools/path"
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
	// Now we know that the user was asking for something possible with the detected build pack,
	// let's make sure that build pack is properly compatible with this project
	issues := pack.CheckCompatibility()
	if len(issues) != 0 {
		cli.Logf("This %s project has some issues...", pack.Name)
		cli.LogBulletList("-", issues)
		cli.Fatal()
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
		return false
	}
	context.IncrementBuildNumber()
	BuildDockerfile(feature, context, appInfo)
	if file.Exists("Dockerfile") {
		cli.Logf("INFO: Your local Dockerfile is ignored by sous, just so you know")
	}
	df := path.Resolve(context.FilePath("Dockerfile"))
	docker.BuildFile(df, ".", context.DockerTag())
	context.Commit()
	return true
}

func BuildDockerfile(feature *build.Feature, context *build.Context, appInfo *build.AppInfo) {
	df := feature.MakeDockerfile(appInfo)
	AddMetadata(df, context)
	context.SaveFile(df.Render(), "Dockerfile")
}

func RequireDocker() {
	docker.RequireVersion(version.Range(">=1.8.2"))
	docker.RequireDaemon()
}

func RequireGit() {
	git.RequireVersion(version.Range(">=1.9.1"))
	git.RequireRepo()
}

func AddMetadata(d *docker.Dockerfile, c *build.Context) {
	d.Maintainer = c.User
	d.AddLabel("builder.app", "sous")
	d.AddLabel("builder.host", c.Host)
	d.AddLabel("builder.fullhost", c.FullHost)
	d.AddLabel("builder.user", c.User)
	d.AddLabel("source.git.repo", c.Git.CanonicalName())
	d.AddLabel("source.git.revision", c.Git.CommitSHA)
}

func divineTaskHost() string {
	taskHost := os.Getenv("TASK_HOST")
	if taskHost != "" {
		return taskHost
	}
	return docker.GetDockerHost()
}

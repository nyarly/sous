package core

import (
	"fmt"
	"os"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/git"
	"github.com/opentable/sous/tools/path"
	"github.com/opentable/sous/tools/version"
)

func (s *Sous) AssembleTargetContext(targetName string) (Target, *Context, *AppInfo) {
	packs := s.Packs
	p := DetectProjectType(packs)
	if p == nil {
		cli.Fatalf("no buildable project detected")
	}
	pack := CompiledPack{Pack: p}
	target, ok := pack.GetTarget(targetName)
	if !ok {
		cli.Fatalf("The %s build pack does not support %s", pack, targetName)
	}
	// Now we know that the user was asking for something possible with the detected build pack,
	// let's make sure that build pack is properly compatible with this project
	issues := pack.Problems()
	if len(issues) != 0 {
		cli.Logf("This %s project has some issues...", pack)
		cli.LogBulletList("-", issues)
		cli.Fatal()
	}
	appVersion := pack.AppVersion()
	if appVersion == "" {
		appVersion = "unversioned"
	}
	context := GetContext(targetName)
	err := target.Check()
	if err != nil {
		cli.Fatalf("unable to %s %s project: %s", targetName, pack, err)
	}
	context.AppVersion = appVersion
	return target, context, nil
}

func (s *Sous) BuildIfNecessary(target Target, context *Context, appInfo *AppInfo) bool {
	if !context.NeedsBuild() {
		return false
	}
	context.IncrementBuildNumber()
	s.BuildDockerfile(target, context, appInfo)
	if file.Exists("Dockerfile") {
		cli.Logf("INFO: Your local Dockerfile is ignored by sous, just so you know")
	}
	df := path.Resolve(context.FilePath("Dockerfile"))
	localDockerfile := ".SousDockerfile"
	cli.AddCleanupTask(func() error {
		file.Remove(localDockerfile)
		if file.Exists(localDockerfile) {
			return fmt.Errorf("Unable to remove temporary file %s", localDockerfile)
		}
		return nil
	})
	docker.BuildFile(df, localDockerfile, ".", context.DockerTag())
	context.Commit()
	return true
}

func (s *Sous) BuildDockerfile(target Target, context *Context, appInfo *AppInfo) {
	df := target.Dockerfile()
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

func AddMetadata(d *docker.Dockerfile, c *Context) {
	d.Maintainer = c.User
	d.AddLabel("builder.app", "sous")
	d.AddLabel("builder.host", c.Host)
	d.AddLabel("builder.fullhost", c.FullHost)
	d.AddLabel("builder.user", c.User)
	d.AddLabel("source.git.repo", c.Git.CanonicalName())
	d.AddLabel("source.git.revision", c.Git.CommitSHA)
}

func DivineTaskHost() string {
	taskHost := os.Getenv("TASK_HOST")
	if taskHost != "" {
		return taskHost
	}
	return docker.GetDockerHost()
}

package core

import (
	"os"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/git"
	"github.com/opentable/sous/tools/path"
	"github.com/opentable/sous/tools/version"
)

func CheckForProblems(pack Pack) (fatal bool) {
	// Now we know that the user was asking for something possible with the detected build pack,
	// let's make sure that build pack is properly compatible with this project
	issues := pack.Problems()
	warnings, errors := issues.Warnings(), issues.Errors()
	if len(warnings) != 0 {
		cli.LogBulletList("WARNING:", issues.Strings())
	}
	if len(errors) != 0 {
		cli.LogBulletList("ERROR:", errors.Strings())
		cli.Logf("ERROR: Your project cannot be built by Sous until the above errors are rectified")
		return true
	}
	return false
}

func (s *Sous) AssembleTargetContext(targetName string) (Target, *Context) {
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
	if fatal := CheckForProblems(pack.Pack); fatal {
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
	return target, context
}

// BuildIfNecessary usually rebuilds any target if anything of the following
// are true:
//
// - No build is available at all
// - Any files in the working tree have changed
// - Sous has been updated
// - Sous config has changed
//
// However, you may override this behaviour for a specific target by implementing
// the Staler interface: { Stale(*Context) bool }
func (s *Sous) BuildImageIfNecessary(target Target, context *Context) bool {
	if !s.NeedsToBuildNewImage(target, context) {
		return false
	}
	s.BuildImage(target, context)
	return true
}

func (s *Sous) BuildImage(target Target, context *Context) {
	context.IncrementBuildNumber()
	if file.Exists("Dockerfile") {
		cli.Logf("WARNING: Your local Dockerfile is ignored by sous, use `sous dockerfile %s` to see the dockerfile being used here", target.Name())
	}
	dfPath := path.Resolve(context.FilePath("Dockerfile"))
	if prebuilder, ok := target.(PreDockerBuilder); ok {
		prebuilder.PreDockerBuild(context)
		// NB: Always rebuild the Dockerfile after running pre-build, since pre-build
		// may update target state to reflect things like copied file locations etc.
		s.BuildDockerfile(target, context)
	}
	docker.BuildFile(dfPath, ".", context.DockerTag())
	context.Commit()
}

func (s *Sous) BuildDockerfile(target Target, context *Context) *docker.Dockerfile {
	df := target.Dockerfile()
	AddMetadata(df, context)
	context.SaveFile(df.Render(), "Dockerfile")
	return df
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

package core

import (
	"os"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/git"
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

func (s *Sous) TargetContext(targetName string) *TargetContext {
	context := GetContext()
	pack := context.DetectProjectType(s.State.Buildpacks)
	if pack == nil {
		cli.Fatalf("no buildable project detected")
	}
	target := GetTarget(pack, context, targetName)
	err := target.Check()
	if err != nil {
		cli.Fatalf("unable to %s %s project: %s", targetName, pack, err)
	}
	bs := GetBuildState(targetName, context.Git)
	return &TargetContext{
		TargetName: targetName,
		BuildState: bs,
		Buildpack:  pack,
		Context:    context,
		Target:     target,
	}
}

func GetTarget(bp *Buildpack, c *Context, name string) Target {
	switch name {
	default:
		return nil
	case "app":
		return NewAppTarget(bp, c)
	case "compile":
		return NewCompileTarget(bp, c)
	case "test":
		return nil
		//return NewTestTarget(bp)
	}
}

// TODO: Remove this func, just use TargetContext
func (s *Sous) AssembleTargetContext(targetName string) (Target, *Context) {
	tc := s.TargetContext(targetName)
	return tc.Target, tc.Context
}

func RequireDocker() {
	docker.RequireVersion(version.Range(">=1.8.2"))
	docker.RequireDaemon()
}

func RequireGit() {
	git.RequireVersion(version.Range(">=1.9.1"))
	git.RequireRepo()
}

func DivineTaskHost() string {
	taskHost := os.Getenv("TASK_HOST")
	if taskHost != "" {
		return taskHost
	}
	return docker.GetDockerHost()
}

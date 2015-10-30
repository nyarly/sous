package commands

import (
	"flag"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/git"
)

func BuildHelp() string {
	return `sous build detects your project type, and tries to find a matching
OpenTable supported stack to build against. Right now it only supports NodeJS
projects. It builds a docker image, tagged and labelled correctly.

sous build does not have any options yet`
}

var buildFlags = flag.NewFlagSet("build", flag.ExitOnError)
var forceBuild = buildFlags.Bool("force", false, "force a new build, even if sous thinks it's not necessary")

func Build(sous *core.Sous, args []string) {
	buildFlags.Parse(args)
	args = buildFlags.Args()
	targetName := "app"
	if len(args) != 0 {
		targetName = args[0]
	}
	core.RequireGit()
	core.RequireDocker()
	if err := git.AssertCleanWorkingTree(); err != nil {
		cli.Logf("WARNING: Dirty working tree: %s", err)
	}

	target, context := sous.AssembleTargetContext(targetName)

	built, _ := sous.RunTarget(target, context)

	if built {
		cli.Successf("Already built: %s", context.DockerTag())
	}

	//if *forceBuild {
	//	cli.Logf("Forcing new build")
	//	sous.Build(target, context)
	//} else {
	//	if !sous.BuildIfNecessary(target, context) {
	//		cli.Successf("Already built: %s", context.DockerTag())
	//	}
	//}
	name := context.CanonicalPackageName()
	cli.Successf("Successfully built %s v%s as %s", name, context.AppVersion, context.DockerTag())
}

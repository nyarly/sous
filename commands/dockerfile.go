package commands

import (
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/file"
)

func DockerfileHelp() string {
	return `sous dockerfile prints the current dockerfile for this project`
}

func Dockerfile(sous *core.Sous, args []string) {
	targetName := "app"
	if len(args) != 0 {
		targetName = args[0]
	}
	target, context, appInfo := sous.AssembleTargetContext(targetName)
	sous.BuildDockerfile(target, context, appInfo)
	fp := context.FilePath("Dockerfile")
	df, ok := file.ReadString(fp)
	if !ok {
		cli.Fatalf("Unable to read %s", fp)
	}
	cli.Outf(df)
	cli.Success()
}

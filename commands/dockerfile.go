package commands

import (
	"fmt"

	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/path"
)

func DockerfileHelp() string {
	return `sous dockerfile prints the current dockerfile for this project`
}

func Dockerfile(packs []*build.Pack, args []string) {
	target := "build"
	if len(args) != 0 {
		target = args[0]
	}
	feature, context, appInfo := AssembleFeatureContext(target, packs)
	BuildDockerfileIfNecessary(feature, context, appInfo)
	fp := context.FilePath("Dockerfile")
	df, ok := file.ReadString(fp)
	if !ok {
		cli.Fatalf("Unable to read %s", fp)
	}
	cli.Outf(df)
	cli.Success()
}

func BuildPathHelp() string {
	return `sous build-path prints the working directory for this project's build state`
}

func BuildPath(packs []*build.Pack, args []string) {
	_, context, _ := AssembleFeatureContext("build", packs)
	fmt.Println(path.Resolve(path.BaseDir(context.BaseDir())))
	cli.Success()
}

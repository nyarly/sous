package commands

import (
	"flag"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
)

func LsHelp() string {
	return `sous ls lists images, contaners and artifacts created by sous`
}

var lsFlags = flag.NewFlagSet("ls", flag.ExitOnError)

func Ls(sous *core.Sous, args []string) {
	//globalFlag := lsFlags.Bool("g", false, "global: list files in all projects sous has built")
	//lsFlags.Parse(args)
	//global := *globalFlag
	args = lsFlags.Args()
	if len(args) != 0 {
		cli.Fatalf("sous ls does not accept any arguments")
	}
	_, context := sous.AssembleTargetContext("app")
	cli.Outf(" ===> Images")
	images := sous.LsImages(context)
	if len(images) == 0 {
		cli.Logf("  no images for this project")
	}
	for _, image := range images {
		cli.Logf("  %s:%s", image.Name, image.Tag)
	}
	cli.Outf(" ===> Containers")
	containers := sous.LsContainers(context)
	if len(containers) == 0 {
		cli.Logf("  no containers for this project")
	}
	for _, container := range containers {
		cli.Logf("  %s (%s)", container.Name(), container.CID())
	}
	cli.Success()
}

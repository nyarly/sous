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
	sous.LsImages(context)
	cli.Outf(" ===> Containers")
	sous.LsContainers(context)
	cli.Success()
}

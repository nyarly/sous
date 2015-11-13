package commands

import (
	"flag"
	"fmt"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/cmd"
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
	lsImages(sous, context)
	cli.Outf(" ===> Containers")
	lsContainers(sous, context)
	cli.Success()
}

func lsImages(sous *core.Sous, c *core.Context) {
	labelFilter := fmt.Sprintf("label=%s.build.package.name=%s", sous.Config.DockerLabelPrefix, c.CanonicalPackageName())
	results := cmd.Table("docker", "images", "--filter", labelFilter)
	// The first line is just table headers
	if len(results) < 2 {
		return
	}
	results = results[1:]
	for _, row := range results {
		cli.Outf("  %s:%s", row[0], row[1])
	}
}

func lsContainers(sous *core.Sous, c *core.Context) {
	labelFilter := fmt.Sprintf("label=%s.build.package.name=%s", sous.Config.DockerLabelPrefix, c.CanonicalPackageName())
	results := cmd.Table("docker", "ps", "-a", "--filter", labelFilter)
	// The first line is just table headers
	if len(results) < 2 {
		return
	}
	results = results[1:]
	for _, row := range results {
		nameIndex := len(row) - 1
		cli.Outf("  %s (%s)", row[nameIndex], row[0])
	}
}

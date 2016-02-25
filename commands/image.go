package commands

import (
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
)

func ImageHelp() string {
	return `sous image prints the last built image tag for this project`
}

func Image(sous *core.Sous, args []string) {
	target := "app"
	if len(args) != 0 {
		target = args[0]
	}
	tc := sous.TargetContext(target)
	if tc.BuildNumber() == 0 {
		cli.Fatalf("no builds yet")
	}
	cli.Outf(tc.DockerTag())
	cli.Success()
}

package commands

import (
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/ports"
)

func TaskPort(sous *core.Sous, args []string) {
	port0, err := ports.GetFreePort()
	if err != nil {
		cli.Fatalf("Unable to get free port: %s", err)
	}
	cli.Outf("%d", port0)
	cli.Success()
}

func TaskPortHelp() string {
	return "random free port number on your task_host"
}

package commands

import (
	"time"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/deploy"
	"github.com/opentable/sous/tools/cli"
)

func UpdateHelp() string {
	return `sous update updates your local sous config cache`
}

func Update(sous *core.Sous, args []string) {
	key := "last-update-check"
	if err := deploy.Update(); err != nil {
		cli.Fatal()
	}
	deploy.Set(key, time.Now().Format(time.RFC3339))
	cli.Success()
}

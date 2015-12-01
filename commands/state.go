package commands

import (
	"sync"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/deploy"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/singularity"
)

func StateHelp() string {
	return `sous state checks the state of all deploys`
}

func State(sous *core.Sous, args []string) {
	stateDir := getStateDir(args)
	state, err := deploy.Parse(stateDir)
	if err != nil {
		cli.Fatalf("%s", err)
	}
	wg := sync.WaitGroup{}
	results := make(chan []singularity.RequestParent, len(state.Datacentres))
	wg.Add(len(state.Datacentres))
	for _, dc := range state.Datacentres {
		go func(dc *deploy.Datacentre) {
			c := singularity.NewClient(dc.SingularityURL)
			rs, err := c.Requests()
			if err != nil {
				cli.Fatalf("%s", err)
			}
			cli.Logf("%s: %d", dc.SingularityURL, len(rs))
			results <- rs
			wg.Done()
		}(dc)
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	for rs := range results {
		cli.Logf("RESULT! %d", len(rs))
	}
	cli.Success()
}

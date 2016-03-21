package commands

import (
	"fmt"
	"sync"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/deploy"
	"github.com/opentable/sous/tools/cli"
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
	merged, err := state.Merge()
	if err != nil {
		cli.Fatalf("%s", err)
	}
	wg := sync.WaitGroup{}
	results := make(chan DiffResult, len(merged.Datacentres))
	wg.Add(len(merged.Datacentres))
	for name := range merged.Datacentres {
		dc := merged.CompiledDatacentre(name)
		go func(dc deploy.CompiledDatacentre) {
			r := dc.DiffRequests()
			results <- DiffResult{
				Datacentre: dc,
				Diffs:      r,
			}
			wg.Done()
		}(dc)
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	for rs := range results {
		fmt.Printf(" ===> %s diffs (%d)\n", rs.Datacentre.Name, len(rs.Diffs))
		for i, d := range rs.Diffs {
			fmt.Printf("  diff %00d: %s\n", i, d.Desc)
		}
	}
	cli.Success()
}

type DiffResult struct {
	Datacentre deploy.CompiledDatacentre
	Diffs      []deploy.Diff
}

package core

import (
	"os"

	"github.com/opentable/sous/tools/cli"
)

func DetectProjectType(packs []*Pack) *Pack {
	availablePacks := []*Pack{}
	for _, p := range packs {
		err := p.Detect()
		if err != nil {
			continue
		}
		availablePacks = append(availablePacks, p)
	}
	if len(availablePacks) == 0 {
		return nil
	}
	if len(availablePacks) > 1 {
		cli.Fatalf("multiple project types detected")
		os.Exit(1)
	}
	return availablePacks[0]
}

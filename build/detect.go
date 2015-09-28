package build

import (
	"fmt"
	"os"
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
		fmt.Println("no project detected")
		os.Exit(0)
	}
	if len(availablePacks) > 1 {
		fmt.Println("multiple project types detected")
		os.Exit(0)
	}
	return availablePacks[0]
}

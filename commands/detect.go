package commands

import (
	"fmt"
	"os"

	"github.com/opentable/sous/build"
)

func Detect(packs []*build.Pack, args []string) {
	pack := build.DetectProjectType(packs)
	if pack == nil {
		fmt.Println("no sous-compatible project detected")
		os.Exit(1)
	}
	fmt.Printf("Detected a %s project\n", pack.Name)
	context := build.GetContext("detect")
	for name, feature := range pack.Features {
		if _, err := feature.Detect(context); err != nil {
			continue
		}
		fmt.Println("this project supports", name)
	}
	os.Exit(0)
}

func DetectHelp() string {
	return `detect detects available actions for your project, and
tells you how to enable those features`
}

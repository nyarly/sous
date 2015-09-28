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
	fmt.Println("Detected a %s project", pack.Name)
	context := build.GetContext()
	if err, _ := pack.Features.Build.Detect(context); err != nil {
		fmt.Println("this project supports build")
	}
	if err, _ := pack.Features.Test.Detect(context); err != nil {
		fmt.Println("this project supports test")
	}
}

func DetectHelp() string {
	return `detect detects available actions for your project, and
tells you how to enable those features`
}

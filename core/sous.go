package core

import (
	"encoding/json"
	"flag"

	"github.com/opentable/sous/config"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
)

type Sous struct {
	Version, Revision, OS, Arch string
	Packs                       []Pack
	Commands                    map[string]*Command
	cleanupTasks                []func() error
	Flags                       *SousFlags
	flagSet                     *flag.FlagSet
}

type SousFlags struct {
	ForceRebuild, ForceRebuildAll bool
}

type Command struct {
	Func      func(*Sous, []string)
	HelpFunc  func() string
	ShortDesc string
}

var sous *Sous

func NewSous(version, revision, os, arch string, commands map[string]*Command, packs []Pack) *Sous {
	if sous == nil {
		sous = &Sous{
			Version:      version,
			Revision:     revision,
			OS:           os,
			Arch:         arch,
			Packs:        packs,
			Commands:     commands,
			Flags:        &SousFlags{},
			cleanupTasks: []func() error{},
		}
	}
	return sous
}

func (s *Sous) ParseFlags(args []string) []string {
	flagSet := flag.NewFlagSet("sous", flag.ExitOnError)
	rebuild := flagSet.Bool("rebuild", false, "force a rebuild")
	rebuildAll := flagSet.Bool("rebuild-all", false, "force a rebuild of this target plus all dependencies")
	err := flagSet.Parse(args)
	if err != nil {
		cli.Fatalf("%s", err)
	}
	s.Flags = &SousFlags{
		ForceRebuild:    *rebuild,
		ForceRebuildAll: *rebuildAll,
	}
	return flagSet.Args()
}

func (s *Sous) UpdateBaseImage(image string) {
	// First, keep track of which images we are interested in...
	key := "usedBaseImages"
	images := config.Properties()[key]
	var list []string
	if len(images) != 0 {
		json.Unmarshal([]byte(images), &list)
	} else {
		list = []string{}
	}
	if doesNotAppearInList(image, list) {
		list = append(list, image)
	}
	listJSON, err := json.Marshal(list)
	if err != nil {
		cli.Fatalf("Unable to marshal base image list as JSON: %+v; %s", list, err)
	}
	config.Set(key, string(listJSON))
	// Now lets grab the actual image
	docker.Pull(image)
}

func doesNotAppearInList(item string, list []string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}
	return false
}

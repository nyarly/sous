package main

import (
	"github.com/opentable/sous/config"
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/packs/golang"
	"github.com/opentable/sous/packs/nodejs"
)

func BuildPacks(c *config.Config) []core.Pack {
	return []core.Pack{
		nodejs.New(c.Packs.NodeJS),
		golang.New(c.Packs.Go),
	}
}

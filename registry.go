package main

import (
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/packs/nodejs"
)

var buildPacks = []*core.Pack{
	nodejs.Pack,
}

package main

import (
	"github.com/opentable/sous/build"
	"github.com/opentable/sous/packs/nodejs"
)

var buildPacks = []*build.Pack{
	nodejs.Pack,
}

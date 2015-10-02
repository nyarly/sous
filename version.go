package main

import (
	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
)

// These values are set at build time using -ldflags "-X main.Name=Value"
var Version, Branch, CommitSHA, BuildNumber, BuildTimestamp, OS, Arch string

func version(packs []*build.Pack, args []string) {
	cli.Outf("sous version %s %s/%s", Version, OS, Arch)
	cli.Outf("Built commit SHA: %s", CommitSHA)
	cli.Success()
}

func versionHelp() string {
	return "Sous version information"
}

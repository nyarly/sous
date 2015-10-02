package main

import (
	"os"

	"github.com/opentable/sous/tools/cli"
)

func main() {
	loadCommands()
	// this line avoids initialisation loop
	if len(os.Args) < 2 {
		usage()
	}
	command := os.Args[1]
	if c, ok := Sous.Commands[command]; ok {
		c.Func(buildPacks, os.Args[2:])
		cli.Fatalf("Command did not complete correctly")
	}
	cli.Fatalf("Command %s not recognised; try `sous help`", command)
}

func usage() {
	cli.Fatalf("usage: sous COMMAND; try `sous help`")
}

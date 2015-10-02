package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
)

func help(packs []*build.Pack, args []string) {
	if len(args) != 0 {
		command := args[0]
		if c, ok := Sous.Commands[command]; ok {
			if c.HelpFunc != nil {
				fmt.Println(c.HelpFunc())
				os.Exit(0)
			}
			cli.Fatalf("Command %s does not have any help yet.", command)
		}
		cli.Fatalf("There is no command called %s; try `sous help`\n", command)
	}
	cli.Outf(`Sous is your personal sous chef for engineering tasks.
																																It can help with building, configuring, and deploying
																																your code for OpenTable's Mesos Platform.

																																Commands:`)

	printCommands()
	cli.Outf("")
	cli.Outf("Tip: for help with any command, use `sous help <COMMAND>`")
	cli.Success()
}

func printCommands() {
	commandNames := make([]string, len(Sous.Commands))
	i := 0
	for n, _ := range Sous.Commands {
		commandNames[i] = n
		i++
	}
	sort.Strings(commandNames)
	for _, name := range commandNames {
		fmt.Printf("\t%s\t%s\n", name, Sous.Commands[name].ShortDesc)
	}
}

func helphelp() string {
	return "Help: /verb/ To give assistance to; aid"
}

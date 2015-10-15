package commands

import (
	"fmt"
	"os"
	"sort"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
)

func Help(sous *core.Sous, args []string) {
	if len(args) != 0 {
		command := args[0]
		if c, ok := sous.Commands[command]; ok {
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

	printCommands(sous)
	cli.Outf("")
	cli.Outf("Tip: for help with any command, use `sous help <COMMAND>`")
	cli.Success()
}

func printCommands(sous *core.Sous) {
	commandNames := make([]string, len(sous.Commands))
	i := 0
	for n := range sous.Commands {
		commandNames[i] = n
		i++
	}
	sort.Strings(commandNames)
	for _, name := range commandNames {
		fmt.Printf("\t%s\t%s\n", name, sous.Commands[name].ShortDesc)
	}
}

func HelpHelp() string {
	return "Help: /verb/ To give assistance to; aid"
}

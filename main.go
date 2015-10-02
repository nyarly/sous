package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/opentable/sous/build"
	"github.com/opentable/sous/commands"
	"github.com/opentable/sous/tools/cli"
)

type SousCommand struct {
	Func      func(packs []*build.Pack, args []string)
	HelpFunc  func() string
	ShortDesc string
}

// These values are set at build time using -ldflags "-X main.Name=Value"
var Version, Branch, CommitSHA, BuildNumber, BuildTimestamp, OS, Arch string

var Sous = struct {
	Commands map[string]SousCommand
}{
	map[string]SousCommand{
		"build":      SousCommand{commands.Build, commands.BuildHelp, "build your project"},
		"push":       SousCommand{commands.Push, commands.PushHelp, "push your project"},
		"run":        SousCommand{commands.Run, commands.RunHelp, "run your project"},
		"dockerfile": SousCommand{commands.Dockerfile, commands.DockerfileHelp, "print current dockerfile"},
		"detect":     SousCommand{commands.Detect, commands.DetectHelp, "detect available actions"},
		"test":       SousCommand{commands.Test, commands.TestHelp, "test your project"},
		"build-path": SousCommand{commands.BuildPath, commands.BuildPathHelp, "build state directory"},
	},
}

func main() {
	// this line avoids initialisation loop
	Sous.Commands["help"] = SousCommand{help, helphelp, "show this help"}
	Sous.Commands["version"] = SousCommand{version, versionHelp, "show version info"}
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
	fmt.Println("usage: sous COMMAND; try `sous help`")
	os.Exit(1)
}

func version(packs []*build.Pack, args []string) {
	cli.Outf("sous version %s %s/%s", Version, OS, Arch)
	cli.Outf("Built commit SHA: %s", CommitSHA)
	cli.Success()
}

func versionHelp() string {
	return "Sous version information"
}

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
		fmt.Printf("There is no command called %s; try `sous help`\n", command)
		os.Exit(1)
	}
	fmt.Println(`Sous is your personal sous chef for engineering tasks.
It can help with building, configuring, and deploying
your code for OpenTable's Mesos Platform.

Commands:`)

	printCommands()
	fmt.Println()
	fmt.Println("Tip: for help with any command, use `sous help <COMMAND>`")
	os.Exit(0)
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

package core

import "flag"

type Command struct {
	Name, Desc, Help string
	Args             []Arg
	Act              func(*Sous, []string) error
	Flags            *flag.FlagSet
	SubCommands      map[string]*Command
}

type Arg struct {
	Name, Default string
	Required      bool
}

var DockerfileCommand = Command{
	Name: "build",
	Desc: "build one of your project's targets (default: app)",
	Args: []Arg{Arg{"target", "app", false}},
	Act: func(*Sous, []string) error {
		return nil
	},
	Flags: flag.NewFlagSet("build", flag.ExitOnError),
}

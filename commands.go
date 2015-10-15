package main

import (
	"github.com/opentable/sous/commands"
	"github.com/opentable/sous/core"
)

func loadCommands() map[string]*core.Command {
	return map[string]*core.Command{
		"build": {
			commands.Build, commands.BuildHelp,
			"build your project"},

		"push": {
			commands.Push, commands.PushHelp,
			"push your project"},

		"run": {
			commands.Run, commands.RunHelp,
			"run your project"},

		"contracts": {
			commands.Contracts, commands.ContractsHelp,
			"check project against platform contracts",
		},

		"logs": {
			commands.Logs, commands.LogsHelp,
			"view stdout and stderr from containers",
		},

		"dockerfile": {
			commands.Dockerfile, commands.DockerfileHelp,
			"print current dockerfile"},

		"image": {
			commands.Image, commands.ImageHelp,
			"print last built docker image tag"},

		"update": {
			commands.Update, commands.UpdateHelp,
			"update sous config",
		},

		"detect": {
			commands.Detect, commands.DetectHelp,
			"detect available actions"},

		"test": {
			commands.Test, commands.TestHelp,
			"test your project"},

		"build-path": {
			commands.BuildPath, commands.BuildPathHelp,
			"build state directory"},

		"help": {
			commands.Help, commands.HelpHelp,
			"show this help"},

		"version": {
			commands.Version, commands.VersionHelp,
			"show version info"},

		"config": {
			commands.Config, commands.ConfigHelp,
			"get/set config properties"},
	}
}

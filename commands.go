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

		"clean": {
			commands.Clean, commands.CleanHelp,
			"delete your project's containers and images"},

		"dockerfile": {
			commands.Dockerfile, commands.DockerfileHelp,
			"print current dockerfile"},

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

		"ls": {
			commands.Ls, commands.LsHelp,
			"list images, containers and other artifacts",
		},

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

		"stamp": {
			commands.Stamp, commands.StampHelp,
			"stamp labels onto docker images"},

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

package commands

import (
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
)

func CleanHelp() string {
	return `sous clean removes all containers and images associated with this project`
}

func Clean(sous *core.Sous, args []string) {
	_, context := sous.AssembleTargetContext("app")
	cleanContainersSucceeded := cleanContainers(sous, context)
	cleanImagesSucceeded := cleanImages(sous, context)
	if cleanContainersSucceeded && cleanImagesSucceeded {
		cli.Success()
	}
	cli.Fatal()
}

func cleanImages(sous *core.Sous, context *core.Context) bool {
	success := true
	for _, i := range sous.LsImages(context) {
		if err := i.Remove(); err != nil {
			cli.Logf("Failed to remove image %s:%s", i.Name, i.Tag)
			success = false
		} else {
			cli.Logf("Removed image %s:%s", i.Name, i.Tag)
		}
	}
	return success
}

func cleanContainers(sous *core.Sous, context *core.Context) bool {
	success := true
	for _, c := range sous.LsContainers(context) {
		if err := c.ForceRemove(); err != nil {
			cli.Logf("Failed to remove container %s (%s)", c.Name(), c.CID())
			success = false
		} else {
			cli.Logf("Removed container %s (%s)", c.Name(), c.CID())
		}
	}
	return success
}

package commands

import (
	"strings"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
)

func StampHelp() string {
	return `sous stamp applies labels to a built app image`
}

func Stamp(sous *core.Sous, args []string) {
	target := "app"
	if len(args) == 0 {
		cli.Fatalf("sous stamp requires at least one argument (a docker label)")
	}
	tc := sous.TargetContext(target)
	if tc.BuildNumber() == 0 {
		cli.Fatalf("no builds yet; sous stamp operates on your last successful build of the app target")
	}

	tag := tc.DockerTag()
	run := docker.NewRun(tag)
	run.AddLabels(parseLabels(args))
	run.StdoutFile = "/dev/null"
	run.StderrFile = "/dev/null"
	container, err := run.Background().Start()
	if err != nil {
		cli.Fatalf("Failed to start container for labeling: %s", err)
	}
	if err := container.KillIfRunning(); err != nil {
		cli.Fatalf("Failed to kill labelling container %s: %s", container, err)
	}
	cid := container.CID()
	if err := docker.Commit(cid, tag); err != nil {
		cli.Fatalf("Failed to commit labelled container %s: %s", container, err)
	}
	cli.Successf("Successfully added labels to %s; remember to push.", tag)
}

func parseLabels(args []string) map[string]string {
	labels := make(map[string]string, len(args))
	for _, a := range args {
		parts := strings.SplitN(a, "=", 2)
		if len(parts) == 2 {
			labels[parts[0]] = parts[1]
		} else {
			labels[a] = ""
		}
	}
	return labels
}

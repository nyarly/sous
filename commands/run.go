package commands

import (
	"net"
	"net/url"
	"os"
	"strconv"

	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/ports"
)

func RunHelp() string {
	return `sous run your project (building first if necessary)`
}

func Run(packs []*build.Pack, args []string) {
	target := "build"
	if len(args) != 0 {
		target = args[0]
	}
	RequireGit()
	RequireDocker()

	feature, context, appInfo := AssembleFeatureContext(target, packs)
	if !BuildIfNecessary(feature, context, appInfo) {
		cli.Logf("No changes since last build, running %s", context.DockerTag())
	}

	dr := docker.NewRun(context.DockerTag())
	port0, err := ports.GetFreePort()
	if err != nil {
		cli.Fatalf("Unable to get free port: %s", err)
	}
	dr.AddEnv("PORT0", strconv.Itoa(port0))
	dr.AddEnv("TASK_HOST", divineTaskHost())
	if code := dr.ExitCode(); code != 0 {
		cli.Fatalf("Run failed with exit code %d", code)
	}
	cli.Success()
}

func divineTaskHost() string {
	taskHost := os.Getenv("TASK_HOST")
	if taskHost != "" {
		return taskHost
	}
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		return "localhost"
	}
	u, err := url.Parse(dockerHost)
	if err != nil {
		return "localhost" // Giving up
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "localhost"
	}
	return host
}

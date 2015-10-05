package commands

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/dockermachine"
	"github.com/opentable/sous/tools/ports"
)

func ContractsHelp() string {
	return `sous contracts tests your project conforms to necessary contracts to run successfully on the OpenTable Mesos platform.`
}

func Contracts(packs []*build.Pack, args []string) {
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

	if !dockermachine.Installed() {
		cli.Fatalf("Contracts currently requires a Mac with docker-machine")
	}
	runningVMs := dockermachine.RunningVMs()
	vms := dockermachine.VMs()
	activeVM := os.Getenv("DOCKER_MACHINE_NAME")
	if activeVM == "" {
		if len(runningVMs) == 1 {
			cli.Logf("No active docker machine.")
			cli.Fatalf(`Tip: eval "$(docker-machine env %s)`, activeVM)
		}
		if len(runningVMs) == 0 {
			cli.Logf("No running docker machine.")
			cli.Fatalf(`Tip: docker-machine start %s && eval "$(docker-machine env %s)"`, vms[0], vms[0])
		}
	}
	var ip string
	for _, vm := range runningVMs {
		if vm == activeVM {
			ip = dockermachine.HostIP(vm)
		}
	}
	if ip == "" {
		cli.Logf("Unable to get IP for docker machine %s", activeVM)
		cli.Fatalf("Tip: docker-machine start %s", activeVM)
	}
	dr := docker.NewRun(context.DockerTag())
	port0, err := ports.GetFreePort()
	if err != nil {
		cli.Fatalf("Unable to get free port: %s", err)
	}
	dr.AddEnv("PORT0", strconv.Itoa(port0))
	dr.AddEnv("TASK_HOST", ip)

	go func() {
		dr.ExitCode()
	}()

	var result *http.Response
	for i := 0; i < 10; i++ {
		result, err = http.Get(fmt.Sprintf("http://%s:%d/health"))
		if err == nil {
			if result.StatusCode == 200 {
				cli.Successf("Success! /health endpoint returned 200")
			}
		}
		time.Sleep(time.Second)
	}

	cli.Fatalf("Contract failed: Service did not responde to HTTP GET /health on port %d within 10 seconds", port0)
}

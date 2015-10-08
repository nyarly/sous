package commands

import (
	"flag"
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

var flags = flag.NewFlagSet("contracts", flag.ContinueOnError)

var timeoutFlag = flags.Duration("timeout", 10*time.Second, "per-contract timeout")

func ContractsHelp() string {
	return `sous contracts tests your project conforms to necessary contracts to run successfully on the OpenTable Mesos platform.`
}

func Contracts(packs []*build.Pack, args []string) {
	flags.Parse(args)
	args = flags.Args()
	timeout := *timeoutFlag
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

	within(timeout, fmt.Sprintf("listens on PORT0 (=%d) - Must respond to :%d/ with any HTTP response", port0, port0), func() bool {
		_, err := http.Get(fmt.Sprintf("http://%s:%d/", ip, port0))
		return err == nil
	})
	within(timeout, "GET /health returns 200", func() bool {
		result, err := http.Get(fmt.Sprintf("http://%s:%d/health", ip, port0))
		return err == nil && result.StatusCode == 200
	})
	cli.Success()
}

func within(d time.Duration, what string, f func() bool) bool {
	start := time.Now()
	end := start.Add(d)
	p := cli.BeginProgress(fmt.Sprintf("Checking it %s", what))
	for {
		if f() {
			p.Done("Success!")
			return true
		}
		if time.Now().After(end) {
			break
		}
		p.Increment()
		time.Sleep(time.Second)
	}
	p.Done(fmt.Sprintf("Timeout; failed to %s within %s", what, d))
	return false
}

package commands

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
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

	taskHost := divineTaskHost()
	port0, err := ports.GetFreePort()
	if err != nil {
		cli.Fatalf("Unable to get free port: %s", err)
	}

	dr := docker.NewRun(context.DockerTag())
	dr.AddEnv("PORT0", strconv.Itoa(port0))
	dr.AddEnv("TASK_HOST", taskHost)
	dr.StdoutFile = context.FilePath("stdout")
	dr.StderrFile = context.FilePath("stderr")
	container, err := dr.Background().Start()
	if err != nil {
		cli.Fatalf("Unable to start container: %s", err)
	}

	failed := 0
	failed += within(timeout, fmt.Sprintf("listens on http://$TASK_HOST:$PORT0 (=http://%s:%d) - Must respond with any HTTP response", taskHost, port0), func() bool {
		_, err := http.Get(fmt.Sprintf("http://%s:%d/", taskHost, port0))
		return err == nil
	})
	failed += within(timeout, "GET /health returns 200", func() bool {
		result, err := http.Get(fmt.Sprintf("http://%s:%d/health", taskHost, port0))
		return err == nil && result.StatusCode == 200
	})

	if err := container.Kill(); err != nil {
		cli.Fatalf("Unable to stop container: %s", err)
	}

	if failed == 0 {
		cli.Success()
	}

	cli.Fatalf("%d contracts failed.", failed)
}

func within(d time.Duration, what string, f func() bool) int {
	start := time.Now()
	end := start.Add(d)
	p := cli.BeginProgress(fmt.Sprintf("Checking it %s", what))
	for {
		if f() {
			p.Done("Success!")
			return 0
		}
		if time.Now().After(end) {
			break
		}
		p.Increment()
		time.Sleep(time.Second)
	}
	p.Done(fmt.Sprintf("Timeout; failed to %s within %s", what, d))
	return 1
}

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

type Contract struct {
	Name, Desc string
	Premise    func(*docker.Run) bool
}

var theContracts = []Contract{
	{
		Name: "Listening",
		Desc: "listens for HTTP traffic at http://$TASK_HOST:$PORT0 where $TASK_HOST and $PORT0 are environment variables containing a valid hostname and a valid, free TCP port number; responds to GET / with any HTTP response code",
		Premise: func(run *docker.Run) bool {
			taskHost := run.Env["TASK_HOST"]
			port0 := run.Env["PORT0"]
			result, err := http.Get(fmt.Sprintf("http://%s:%d/", taskHost, port0))
			return err == nil && result.StatusCode > 0
		},
	},
	{
		Name: "Health Endpoint",
		Desc: "responds to HTTP GET /health with HTTP Status Code 200",
		Premise: func(run *docker.Run) bool {
			taskHost := run.Env["TASK_HOST"]
			port0 := run.Env["PORT0"]
			result, err := http.Get(fmt.Sprintf("http://%s:%d/health", taskHost, port0))
			return err == nil && result.StatusCode == 200
		},
	},
}

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

	cli.Logf("=> Running Contracts")
	cli.Logf(`=> TIP: Open another terminal in this directory and type "sous logs -f"`)

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
	for _, c := range theContracts {
		cli.Logf(`=> Checking Contract "%s"`, c.Name)
		cli.Logf(`Description: %s`, c.Desc)
		failed += within(timeout, func() bool {
			return c.Premise(dr)
		})
	}

	if err := container.Kill(); err != nil {
		cli.Logf("WARNING: Unable to stop container %s: %s", container.CID, err)
	}

	if failed != 0 {
		cli.Fatalf("%d contracts failed.", failed)
	}

	cli.Success()
}

func within(d time.Duration, f func() bool) int {
	start := time.Now()
	end := start.Add(d)
	p := cli.BeginProgress("Polling")
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
	p.Done("Timeout")
	return 1
}

package commands

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/ports"
)

var contractsFlags = flag.NewFlagSet("contracts", flag.ExitOnError)

var timeoutFlag = contractsFlags.Duration("timeout", 10*time.Second, "per-contract timeout")

type Contract struct {
	Name    string
	Desc    func(*docker.Run) string
	Tips    func(*docker.Run) []string
	Premise func(*docker.Run) bool
}

var theContracts = []Contract{
	{
		Name: "Listening on http://TASK_HOST:PORT0",
		Desc: func(run *docker.Run) string {
			return fmt.Sprintf("Your app should respond to GET http://TASK_HOST:PORT0/ with any HTTP response code")
		},
		Tips: func(run *docker.Run) []string {
			host, port0 := run.Env["TASK_HOST"], run.Env["PORT0"]
			return []string{
				fmt.Sprintf("TASK_HOST and PORT0 are environment variables set by the docker run command."),
				fmt.Sprintf("For this particular run they are set as: TASK_HOST=%s and PORT0=%s", host, port0),
				fmt.Sprintf("So your app should be listening on http://%s:%s/", host, port0),
			}
		},
		Premise: func(run *docker.Run) bool {
			taskHost := run.Env["TASK_HOST"]
			port0 := run.Env["PORT0"]
			result, err := http.Get(fmt.Sprintf("http://%s:%d/", taskHost, port0))
			return err == nil && result.StatusCode > 0
		},
	},
	{
		Name: "Health Endpoint",
		Desc: func(run *docker.Run) string {
			return "responds to GET /health with HTTP Status Code 200"
		},
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

func Contracts(sous *core.Sous, args []string) {
	contractsFlags.Parse(args)
	args = contractsFlags.Args()
	timeout := *timeoutFlag
	targetName := "app"
	if len(args) != 0 {
		targetName = args[0]
	}
	core.RequireGit()
	core.RequireDocker()

	target, context := sous.AssembleTargetContext(targetName)

	sous.RunTarget(target, context)

	cli.Logf("=> Running Contracts")
	cli.Logf(`=> **TIP:** Open another terminal in this directory and type **sous logs -f**`)

	taskHost := core.DivineTaskHost()
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
	cli.AddCleanupTask(func() error {
		return container.KillIfRunning()
	})

	failed := 0
	for _, c := range theContracts {
		cli.Logf(`===> CHECKING CONTRACT: "%s"`, c.Name)
		cli.Logf(`===> Description: %s`, c.Desc(dr))
		if c.Tips != nil {
			cli.Logf("===> **TIPS for this contract:**")
			cli.LogBulletList("     -", c.Tips(dr))
		}
		failed += within(timeout, func() bool {
			return c.Premise(dr)
		})
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

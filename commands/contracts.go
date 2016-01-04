package commands

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/deploy"
	"github.com/opentable/sous/tools"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/cmd"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/yaml"
)

var contractsFlags = flag.NewFlagSet("contracts", flag.ExitOnError)

var timeoutFlag = contractsFlags.Duration("timeout", 10*time.Second, "per-contract timeout")
var dockerImage = contractsFlags.String("image", "", "run contracts against a pre-built Docker image")

func ContractsHelp() string {
	return `sous contracts tests your project conforms to necessary contracts to run successfully on the OpenTable Mesos platform.`
}

func Contracts(sous *core.Sous, args []string) {
	contractsFlags.Parse(args)
	args = contractsFlags.Args()
	if len(args) != 1 {
		cli.Fatalf("You must supply one argument: the docker image you want to run contracts against.")
	}

	image := args[0]
	if !docker.ImageExists(image) {
		cli.Logf("Image %q not found locally; pulling...")
		docker.Pull(image)
	}

	state, err := deploy.Parse(".")
	if err != nil {
		cli.Fatalf("Unable to parse state: %s", err)
	}

	contracts := state.Contracts

	initialValues := map[string]string{
		"Image": image,
	}

	for _, name := range state.ContractDefs["http-service"] {
		contract, ok := contracts[name]
		if !ok {
			cli.Fatalf("Contract %q not defined but was listed in http-service contract defs.")
		}
		run := NewContractRun(contract, initialValues)
		if err := run.Execute(); err != nil {
			cli.Fatalf("Contract %q failed: %s", name, err)
		}
	}

	cli.Success()
}

func (r *ContractRun) Execute() error {
	c := r.Contract
	cli.Logf("Running contract: %s", c.Name)
	for _, serverName := range c.StartServers {
		if err := r.StartServer(serverName); err != nil {
			return err
		}
	}
	return nil
}

func (r *ContractRun) StartServer(serverName string) error {
	c := r.Contract
	s, ok := c.Servers[serverName]
	if !ok {
		return fmt.Errorf("Contract %q specifies %s in StartServers, but no server with that name exists", c.Name, serverName)
	}
	var err error
	var startedServer *StartedServer
	startedServer, values, err := startServer(s, r.GlobalValues)
	if err != nil {
		return err
	}
	cli.Logf("Started.")
	cli.AddCleanupTask(func() error {
		cli.Logf("Stopping container %s", startedServer.Container)
		return startedServer.Container.KillIfRunning()
	})
	r.GlobalValues = values
	r.Servers[serverName] = startedServer
	return nil
}

type ContractRun struct {
	Contract     deploy.Contract
	GlobalValues map[string]string
	Servers      map[string]*StartedServer
}

func NewContractRun(contract deploy.Contract, initialValues map[string]string) *ContractRun {
	if initialValues == nil {
		initialValues = map[string]string{}
	}
	return &ContractRun{
		Contract:     contract,
		GlobalValues: initialValues,
		Servers:      map[string]*StartedServer{},
	}
}

type StartedServer struct {
	deploy.TestServer
	Container docker.Container
}

func startServer(s deploy.TestServer, values map[string]string) (*StartedServer, map[string]string, error) {
	// 1. Calculate s.Values (Inputs)
	// 2. Render yaml.Marshal(s) to $template
	// 3. Render $template using Go text templating, injecting values (1.)
	// 4. Parse rendered template as YAML to create the runnable server def
	// 5. Construct a docker.Run based on (4.)
	//dockerRun := docker.NewRun(s.Docker.Image)
	// == 1. Calculate values
	for k, v := range s.DefaultValues {
		// Don't use default value if we already have that value in the global agglomeration.
		if _, ok := values[k]; ok {
			continue
		}
		v = tools.TrimWhitespace(v)
		if !strings.HasPrefix(v, "$(") {
			values[k] = v
			continue
		}
		v = trimPrefixAndSuffix(v, "$(", ")")
		result := cmd.Stdout("/bin/sh", "-c", v)
		values[k] = result
		cli.Logf("DEBUG>>> %q => %q", v, result)
	}
	// 2&3&4:
	err := yaml.InjectTemplatePipeline(s, &s, values)
	// 5. Construct a docker.Run based on the fleshed-out server def
	run, err := makeDockerRun(s)
	if err != nil {
		return nil, nil, err
	}

	// ... now start the server ...
	cli.Logf("Starting server %q (%s)", s.Name, s.Docker.Image)
	run.StdoutFile = "/dev/null"
	run.StderrFile = "/dev/null"
	container, err := run.Start()
	if err != nil {
		return nil, nil, err
	}
	startedServer := &StartedServer{s, container}

	return startedServer, values, nil
}

func makeDockerRun(s deploy.TestServer) (*docker.Run, error) {
	d := s.Docker
	run := docker.NewRun(d.Image)
	for k, v := range d.Env {
		run.AddEnv(k, v)
	}
	run.Args = d.Args
	return run, nil
}

func trimPrefixAndSuffix(s, prefix, suffix string) string {
	return strings.TrimSuffix(strings.TrimPrefix(s, prefix), suffix)
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

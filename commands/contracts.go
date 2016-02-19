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
	"github.com/opentable/sous/tools/version"
	"github.com/opentable/sous/tools/yaml"
)

var (
	contractsFlags = flag.NewFlagSet("contracts", flag.ExitOnError)
	timeoutFlag    = contractsFlags.Duration("timeout", 10*time.Second, "per-contract timeout")
	dockerImage    = contractsFlags.String("image", "", "run contracts against a pre-built Docker image")
	contractName   = contractsFlags.String("contract", "", "run a single, named contract")
	checkNumber    = contractsFlags.Int("check", 0, "run a single check within the named contract (only available in conjunction with -contract)")
	listContracts  = contractsFlags.Bool("list", false, "list all contracts")
	listChecks     = contractsFlags.Bool("list-checks", false, "list all checks")
	selfTest       = contractsFlags.Bool("self-test", false, "run contract self tests")
)

func ContractsHelp() string {
	return `sous contracts tests your project conform to contracts specified by sous state`
}

func Contracts(sous *core.Sous, args []string) {
	contractsFlags.Parse(args)
	args = contractsFlags.Args()
	if len(args) != 0 {
		cli.Logf("Usage of sous contracts:")
		contractsFlags.PrintDefaults()
		cli.Fatal()
	}
	if *contractName != "" {
		if _, ok := sous.State.Contracts[*contractName]; !ok {
			cli.Fatalf("Contract %q is not defined.", *contractName)
		}
	}
	handleSelfTestFlags(sous.State, *selfTest, *contractName)
	handleListFlags(sous.State.Contracts, *listContracts, *listChecks, *contractName)

	getInitialValues := func() map[string]string {
		return map[string]string{"Image": *dockerImage}
	}

	cc := NewConfiguredContracts(sous.State, getInitialValues)

	contract := *contractName
	check := *checkNumber
	if check != 0 {
		if contract == "" {
			cli.Fatalf("you specified -check but not -contract")
		}
		singleContract, ok := cc.Contracts[contract]
		if !ok {
			cli.Fatalf("Contract %q does not exist, try `contracts -list`", contract)
		}
		if len(singleContract.Checks) < check {
			cli.Fatalf("Contract %q has %d checks, you asked to run %d",
				contract, len(singleContract.Checks), check)
		}
	}

	docker.RequireVersion(version.Range("^1.8.3"))
	docker.RequireDaemon()

	// If a docker image is not passed in, fall back to normal
	// sous project context to generate an image if necessary.
	if *dockerImage == "" {
		t, c := sous.AssembleTargetContext("app")
		if yes, reason := sous.NeedsToBuildNewImage(t, c, false); yes {
			cli.Logf("Building new image because %s", reason)
			sous.RunTarget(t, c)
		}
		*dockerImage = c.DockerTag()
	}

	if !docker.ImageExists(*dockerImage) {
		cli.Logf("Image %q not found locally; pulling...", *dockerImage)
		docker.Pull(*dockerImage)
	}

	var err error
	if check != 0 {
		err = cc.RunSingleCheck(contract, check)
	} else if contract != "" {
		err = cc.RunSingleContract(contract)
	} else {
		err = cc.RunContractsForKind("http-service")
	}

	if err != nil {
		cli.Fatalf("%s", err)
	}

	cli.Success()
}

type ConfiguredContracts struct {
	Contracts     deploy.Contracts
	ContractDefs  map[string][]string
	InitialValues func() map[string]string
}

func NewConfiguredContracts(state *deploy.State, initialValues func() map[string]string) ConfiguredContracts {
	if err := state.Contracts.Validate(); err != nil {
		cli.Fatalf("Unable to run: %s", err)
	}
	return ConfiguredContracts{state.Contracts, state.ContractDefs, initialValues}
}

func handleSelfTestFlags(state *deploy.State, selfTest bool, singleContract string) {
	if !selfTest {
		return
	}
	// pass in nil here, since we want to make sure it crashes
	// if the test does not pass initialvalues.
	if singleContract != "" {
		if err := RunSingleSelfTest(state.Contracts[singleContract]); err != nil {
			cli.Fatalf("%s", err)
		}
		cli.Success()
	}
	for _, contract := range state.Contracts {
		if err := RunSingleSelfTest(contract); err != nil {
			cli.Fatalf("%s", err)
		}
	}
	cli.Success()
}

func RunSingleSelfTest(contract deploy.Contract) error {
	c := contract
	if len(c.SelfTest.CheckTests) == 0 {
		return fmt.Errorf("contract %q has no check tests", contractName)
	}
	for i, check := range c.Checks {

		ct := c.SelfTest.CheckTests[i]

		cli.Logf("DEBUG>>>>>>>>>>>>>>>>>> % +v", c)

		failRun := NewContractRun(c, map[string]string{"Image": ct.TestImages.Fail})
		cli.Logf(" ==> Testing check FAILS %d (%q) in contract %q fails for image %s",
			i, check.Name, c.Name, ct.TestImages.Fail)
		if err := failRun.ExecuteUpToCheck(i + 1); err == nil {
			return fmt.Errorf("expected image %s to fail check %d (%q) in contract %q, but it passed.",
				ct.TestImages.Fail, i, check.Name, c.Name)
		}
		cli.Logf(" ==> Successfully failed.")

		cli.Logf("DEBUG>>>>>>>>>>>>>>>>>> % +v", c)

		passRun := NewContractRun(c, map[string]string{"Image": ct.TestImages.Pass})
		cli.Logf(" ==> Testing check PASSES %d (%q) in contract %q for image %s",
			i, check.Name, c.Name, ct.TestImages.Pass)
		if err := passRun.ExecuteUpToCheck(i + 1); err != nil {
			return fmt.Errorf("expected image %s to pass check %d (%q) in contract %q, but it failed with error: %s",
				ct.TestImages.Pass, i, check.Name, c.Name, err)
		}
		cli.Logf(" ==> Successfully passed.")
	}
	return nil
}

func (cc ConfiguredContracts) RunContractsForKind(kind string) error {
	for _, name := range cc.ContractDefs[kind] {
		if err := cc.RunSingleContract(name); err != nil {
			return fmt.Errorf("running contracts for %q; %s", kind, err)
		}
	}
	return nil
}

func (cc ConfiguredContracts) RunSingleContract(name string) error {
	initialValues := cc.InitialValues()
	contract, ok := cc.Contracts[name]
	if !ok {
		return fmt.Errorf("contract %q not found.", name)
	}
	run := NewContractRun(contract, initialValues)
	if err := run.Execute(); err != nil {
		return fmt.Errorf("contract %q failed: %s", contract.Name, err)
	}
	return nil
}

func (cc ConfiguredContracts) RunSingleCheck(name string, check int) error {
	initialValues := cc.InitialValues()
	contract, ok := cc.Contracts[name]
	if !ok {
		return fmt.Errorf("contract %q not found.", name)
	}
	run := NewContractRun(contract, initialValues)
	if err := run.ExecuteUpToCheck(check); err != nil {
		return fmt.Errorf("contract %q failed: %s", contract.Name, err)
	}
	return nil
}

// ContractRun is a single execution of a contract. It is also the struct passed
// in when resolving templated values in the contract definition YAML.
type ContractRun struct {
	Contract deploy.Contract
	// GlobalValues is shared between all servers started by the contract.
	// Once written, no item in GlobalValues should ever be changed.
	GlobalValues map[string]string
	// Values contains the resolved values for a specific contract. Typically
	// these are defined as Go text templated values in the contract definition
	// YAML.
	Values map[string]string
	// Servers contains all the started servers for this contract run.
	Servers       map[string]*StartedServer
	Preconditions []string
	Checks        []string
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

// ExecuteUpToCheck executes the first n checks. It is mainly used for
// testing.
func (r *ContractRun) ExecuteUpToCheck(n int) error {
	c := r.Contract
	cli.Logf("** ==> Running contract: %q**", c.Name)

	// First make sure all the necessary servers are started, in the correct order.
	for _, serverName := range c.StartServers {
		if err := r.StartServer(serverName); err != nil {
			return err
		}
	}

	// Second, resolve templated Values map in the contract, in light of the
	// started servers and everything else inside the ContractRun at this point.
	var values map[string]string
	if err := yaml.InjectTemplatePipeline(c.Values, &values, r); err != nil {
		return err
	}
	r.Values = values

	// Third, resolve all the other templated values in the contract using the special
	// Values map. (This is done in 2 stages to make the contracts significantly more
	// readable.
	d := c
	if err := yaml.InjectTemplatePipeline(c, &d, values); err != nil {
		return err
	}

	// Next execute all the precondition checks to ensure we can meaningfully
	// run the main contract checks.
	for _, p := range d.Preconditions {
		if err := ExecuteCheck(p); err != nil {
			return fmt.Errorf("Precondition %q failed: %s", p.String(), err)
		}
		cli.Verbosef(" ==> Precondition **%s** passed.", p)
	}

	// Finally run the actual contract checks.
	for i := 0; i < n; i++ {
		check := d.Checks[i]
		if err := ExecuteCheck(check); err != nil {
			return fmt.Errorf("     check failed: %s; %s", check, err)
		}
		cli.Logf("     check passed: **%s**", check)
	}

	return nil
}

// Execute executes the entire contract.
func (r *ContractRun) Execute() error {
	return r.ExecuteUpToCheck(len(r.Contract.Checks))
}

func ExecuteCheck(c deploy.Check, progressTitle ...string) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Second
	}
	return c.Execute()
}

func (r *ContractRun) StartServer(serverName string) error {
	c := r.Contract
	s, ok := c.Servers[serverName]
	if !ok {
		return fmt.Errorf("Contract %q specifies %s in StartServers, but no server with that name exists", c.Name, serverName)
	}
	resolvedServer, err := r.ResolveServer(s)
	if err != nil {
		return err
	}
	server, err := resolvedServer.Start()
	if err != nil {
		return err
	}
	cli.Verbosef("Started server %q (%s) as %s", serverName, resolvedServer.Docker.Image, server.Container.CID())
	cli.AddCleanupTask(func() error {
		if !server.Container.Running() {
			cli.Verbosef("Not stopping %q container (%s), it had already stopped.", server.ResolvedServer.Name, server.ContainerID)
			return nil
		}
		if err := server.Container.KillIfRunning(); err != nil {
			cli.Logf("Failed to stop %q container (%s)", serverName, server.Container.CID())
		} else {
			cli.Verbosef("Stopped %q container (%s)", serverName, server.Container.CID())
		}
		return err
	})
	r.Servers[serverName] = server
	return nil
}

// ResolvedServer is a *deploy.TestServer whose templated values
// have all been expanded, and is thus ready to be run.
type ResolvedServer deploy.TestServer

type StartedServer struct {
	*ResolvedServer
	// ContainerID is used in contract definitions to address the container.
	ContainerID string
	Container   docker.Container
}

// ResolveServer fleshes out all templated values in the server in the
// context of the current contract run, adding values to the .GlobalValues
// map if they aren't yet set.
func (r *ContractRun) ResolveServer(s deploy.TestServer) (*ResolvedServer, error) {
	cli.Verbosef("Resolving values for server %q", s.Name)
	for k, v := range s.DefaultValues {
		// Don't use default value if we already have that value in the global agglomeration.
		if v, ok := r.GlobalValues[k]; ok {
			cli.Verbosef(" ==> %s=%q (already set)", k, v)
			continue
		}
		v = tools.TrimWhitespace(v)
		if !strings.HasPrefix(v, "$(") {
			r.GlobalValues[k] = v
			cli.Verbosef(" ==> %s=%q", k, v)
			continue
		}
		v = trimPrefixAndSuffix(v, "$(", ")")
		result := cmd.Stdout("/bin/sh", "-c", v)
		r.GlobalValues[k] = result
		cli.Verbosef(" ==> %s=%q (%s)", k, result, v)
	}

	var ss deploy.TestServer
	if err := yaml.InjectTemplatePipeline(s, &ss, r.GlobalValues); err != nil {
		return nil, err
	}

	rs := ResolvedServer(ss)
	return &rs, nil
}

func (s *ResolvedServer) Start() (*StartedServer, error) {
	if !docker.ImageExists(s.Docker.Image) {
		cli.Logf("Image %q does not exist, beginning pull...", s.Docker.Image)
		docker.Pull(s.Docker.Image)
		if !docker.ImageExists(s.Docker.Image) {
			return nil, fmt.Errorf("Docker image %q still missing after pull", s.Docker.Image)
		}
	}
	if !docker.ExactlyOneImageExists(s.Docker.Image) {
		return nil, fmt.Errorf("Docker tag %q points to multiple images", s.Docker.Image)
	}
	run, err := s.MakeDockerRun()
	if err != nil {
		return nil, err
	}
	run.StdoutFile = "/dev/null"
	run.StderrFile = "/dev/null"
	cli.Verbosef("shell> %s", run.CalculatedCommand())
	container, err := run.Start()
	if err != nil {
		return nil, err
	}
	startedServer := &StartedServer{s, container.CID(), container}
	if s.Startup != nil {
		if err := ExecuteCheck(*s.Startup.CompleteWhen, fmt.Sprintf("Waiting for %s server", s.Name)); err != nil {
			return nil, fmt.Errorf("%s failed to start within the timeout (%s): %s", s.Docker.Image, s.Startup.CompleteWhen.Timeout, err)
		}
	}
	return startedServer, nil
}

func (s *ResolvedServer) MakeDockerRun() (*docker.Run, error) {
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

func handleListFlags(contracts deploy.Contracts, listContracts, listChecks bool, contractName string) {
	if listContracts {
		for name, _ := range contracts {
			cli.Outf("%s", name)
		}
		cli.Success()
	}

	if listChecks && contractName == "" {
		for name, c := range contracts {
			cli.Outf("* %s", name)
			for _, check := range c.Checks {
				cli.Outf("  - %s", check)
			}
		}
		cli.Success()
	}
	if listChecks {
		for _, check := range contracts[contractName].Checks {
			cli.Outf("%s", check.Name)
		}
	}
}

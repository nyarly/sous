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
	singleContract = contractsFlags.String("contract", "", "run a single, named contract")
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

	contracts := getContractsInScope(sous.State, *singleContract)

	if *selfTest {
		runSelfTests(contracts)
	}

	handleListFlags(sous.State.Contracts, *listContracts, *listChecks, *singleContract)

	if *dockerImage == "" {
		*dockerImage = getDockerImageFromTargetContext(sous)
	} else {
		if !docker.ImageExists(*dockerImage) {
			cli.Logf("Image %q not found locally; pulling...", *dockerImage)
			docker.Pull(*dockerImage)
		}
	}

	docker.RequireVersion(version.Range("^1.8.3"))
	docker.RequireDaemon()

	validateRunSpecificCheckFlag(*checkNumber, contracts, *dockerImage)

	getInitialValues := func() map[string]string {
		return map[string]string{"Image": *dockerImage}
	}

	failed := false
	for _, contract := range contracts {
		run := NewContractRun(contract, getInitialValues())
		f := run.Execute
		if *checkNumber != 0 {
			f = func() error { return run.ExecuteUpToCheck(*checkNumber) }
		}
		if err := f(); err != nil {
			cli.Warn(" =x=> Contract failed: %s", err)
			failed = true
		}
	}
	if failed {
		cli.Fatalf("Error: One or more contracts failed.")
	}
	cli.Successf("All contracts passed.")
}

func getDockerImageFromTargetContext(sous *core.Sous) string {
	// If a docker image is not passed in, fall back to normal
	// sous project context to generate an image if necessary.
	t, c := sous.AssembleTargetContext("app")
	if yes, reason := sous.NeedsToBuildNewImage(t, c, false); yes {
		cli.Logf("Building new image because %s", reason)
		sous.RunTarget(t, c)
	}
	return c.DockerTag()
}

func getContractsInScope(state *deploy.State, singleContract string) deploy.OrderedContracts {
	if singleContract == "" {
		oc, err := state.ContractsForKind("http-service")
		if err != nil {
			cli.Fatal(err)
		}
		return oc
	}
	c, ok := state.Contracts[singleContract]
	if !ok {
		cli.Fatalf("contract %q is not defined", singleContract)
	}
	return deploy.OrderedContracts{c}
}

func validateRunSpecificCheckFlag(checkNumber int, contracts deploy.OrderedContracts, image string) {
	if checkNumber == 0 {
		return
	}
	if len(contracts) != 1 {
		cli.Fatalf("you specified -check but not -contract")
	}
	contract := contracts[0]
	numChecks := len(contract.Checks)
	if numChecks < checkNumber {
		cli.Fatalf("contract %q has %d checks, you asked to run %d", contract, numChecks, checkNumber)
	}
}

type ConfiguredContracts struct {
	Contracts     deploy.Contracts
	ContractDefs  map[string][]string
	InitialValues func() map[string]string
}

func runSelfTests(contracts deploy.OrderedContracts) {
	failed := false
	for _, c := range contracts {
		if err := RunSingleSelfTest(c); err != nil {
			cli.Logf("%s", err)
			failed = true
		}
	}
	if failed {
		cli.Fatalf("One or more self-tests failed or were missing.")
	}
	cli.Success()
}

func RunSingleSelfTest(contract deploy.Contract) error {
	c := contract
	if len(c.SelfTest.CheckTests) == 0 {
		return fmt.Errorf("contract %q has no check tests", contract.Name)
	}
	cli.Logf("")
	cli.Logf("** ==> Running self-tests for contract %q**", contract.Name)
	cli.Logf("")

	for i, check := range c.Checks {

		checkNum := i + 1

		ct := c.SelfTest.CheckTests[i]

		failRun := NewContractRun(c, map[string]string{"Image": ct.TestImages.Fail})
		cli.Logf(" ==> Testing check %d %q FAILS for image %s",
			checkNum, check.Name, ct.TestImages.Fail)
		if err := failRun.ExecuteUpToCheck(checkNum); err == nil {
			return fmt.Errorf(" ==> TEST FAILED; expected image %s to fail check %d (%q), but it passed.",
				ct.TestImages.Fail, checkNum, check.Name)
		}
		cli.Logf(" ==> TEST PASSED; The check failed correctly.")

		passRun := NewContractRun(c, map[string]string{"Image": ct.TestImages.Pass})
		cli.Logf(" ==> Testing check %d %q PASSES for image %s",
			checkNum, check.Name, ct.TestImages.Pass)
		if err := passRun.ExecuteUpToCheck(checkNum); err != nil {
			return fmt.Errorf(" ==> TEST FAILED; expected image %s to pass check %d (%q), but it failed with error: %s",
				ct.TestImages.Pass, checkNum, check.Name, err)
		}
		cli.Logf(" ==> TEST PASSED; The check succeeded correctly.")
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
// self-testing.
func (r *ContractRun) ExecuteUpToCheck(n int) error {
	c := r.Contract.Clone()
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

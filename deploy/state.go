package deploy

import (
	"fmt"

	"github.com/opentable/sous/config"
)

type State struct {
	config.Config
	EnvironmentDefs EnvDefs
	Datacentres     Datacentres
	Manifests       Manifests
	Contracts       Contracts
}

type EnvDefs map[string]*EnvDef

type EnvDef map[string]*VarDef

type VarDef struct {
	Type       VarType
	Name, Desc string
	Automatic  bool
}

type VarType string

const (
	URL_VARTYPE    = VarType("url")
	INT_VARTYPE    = VarType("int")
	STRING_VARTYPE = VarType("string")
)

type Datacentres map[string]*Datacentre

type Datacentre struct {
	Name, Desc         string
	SingularityURL     string
	DockerRegistryHost string
	Env                DatacentreEnv
}

type DatacentreEnv map[string]string

type Manifests map[string]Manifest

type Manifest struct {
	App         App
	Deployments Deployments
}

type App struct {
	SourceRepo, Owner, Kind string
}

type Deployments map[string]Deployment

type Deployment struct {
	Instance                  Instance
	SourceTag, SourceRevision string
	Environment               map[string]string
}

type Instance struct {
	Count  int
	CPUs   float32
	Memory string
}

type MemorySize string

type Contracts map[string]Contract

type Contract struct {
	Name, Desc    string
	StartServers  []string
	Values        map[string]string
	Servers       map[string]TestServer
	Preconditions []Check
	Checks        []Check
}

type TestServer struct {
	Name, Desc    string
	DefaultValues map[string]string
	Export        []string
	Docker        DockerServer
}

type DockerServer struct {
	Image         string
	Env           map[string]string
	Options, Args []string
}

type GetHTTPAssertion struct {
	URL, ResponseBodyContains, ResponseJSONContains string
	ResponseStatusCode                              int
	AnyResponse                                     bool
}

// Check MUST specify exactly one of GET, Shell, or Contract. If
// more than one of those are specified the check is invalid.
type Check struct {
	Name, Desc string
	// GET must be a URL, or empty if Shell is not empty.
	// The following 4 fields are assertions about
	// the response after getting that URL via HTTP.
	GET                string
	StatusCode         int
	StatusCodeRange    []int
	BodyContainsJSON   interface{}
	BodyContainsString string

	// Shell must be a valid POSIX shell command, or empty if GET is not
	// empty. The command will be executed and the exit code checked
	// against the expected code (note that ints default to zero, so the
	// default case is that we expect a success (0) exit code.
	Shell    string
	ExitCode int
}

// Validate checks that we have a well-formed check.
func (c *Check) Validate() error {
	if c.GET == "" && c.Shell == "" {
		return fmt.Errorf("none of GET, Shell are specified")
	}
	if c.GET != "" && c.Shell != "" {
		return fmt.Errorf("both GET and Shell are specified")
	}
	return nil
}

func (c Check) String() string {
	if c.Name != "" {
		return c.Name
	}
	if c.Shell != "" {
		return c.Shell
	}
	if c.GET != "" {
		return fmt.Sprintf("GET %s", c.GET)
	}
	return "INVALID CHECK"
}

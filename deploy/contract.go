package deploy

import (
	"fmt"
	"time"
)

type Contracts map[string]Contract

type Contract struct {
	Name          string
	StartServers  []string
	Values        map[string]string
	Servers       map[string]TestServer
	Preconditions []Check
	Checks        []Check
}

type TestServer struct {
	Name, Desc    string
	DefaultValues map[string]string
	Startup       *StartupInfo
	Docker        DockerServer
}

type StartupInfo struct {
	Timeout      time.Duration
	CompleteWhen *Check
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
// more than one of those are specified the check is invalid. This
// slightly ugly switching makes the YAML contract definitions
// much more readable, and is easily verifiable.
type Check struct {
	Name       string
	Timeout    time.Duration
	HTTPCheck  `yaml:",inline"`
	ShellCheck `yaml:",inline"`
}

type HTTPCheck struct {
	// GET must be a URL, or empty if Shell is not empty.
	// The following 4 fields are assertions about
	// the response after getting that URL via HTTP.
	GET                string
	StatusCode         int
	StatusCodeRange    []int
	BodyContainsJSON   interface{}
	BodyContainsString string
}

type ShellCheck struct {
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

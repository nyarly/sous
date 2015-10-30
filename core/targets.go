package core

import (
	"fmt"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
)

type Target interface {
	fmt.Stringer
	// Name of the target, as used in command-line operations.
	Name() string
	// GenericDesc is a generic description of the target, applicable to any pack. It is
	// used mostly for help and exposition.
	GenericDesc() string
	// DependsOn lists the direct dependencies of this target. Dependencies listed as "optional" will
	// always be built when available, but if they are not available will be ignored. It is the job
	// of each package under packs/ to correctly define these relationships.
	DependsOn() []Target
	// Desc is a description of what this target does exactly in the context
	// of the pack that owns it. It should be set by the pack when it is initialised.
	Desc() string
	// Check is a function which tries to detect if this target is possible with the
	// current project. If so, it returns a complete *AppInfo, else nil and an error.
	Check() error
	// Dockerfile is the shebang method which writes out a functionally complete *docker.Dockerfile
	// This method is only invoked only once the Detect func has successfully detected target availability.
	// The *AppInfo from Detect is passed in as context.
	Dockerfile() *docker.Dockerfile
}

// Target describes a buildable Docker image that performs a particular task related to building
// testing and deploying the application. Each pack under packs/ will customise its targets for
// the specific jobs that need to be performed for that pack.
type TargetBase struct {
	name,
	genericDesc string
}

func (t *TargetBase) Name() string {
	return t.name
}

func (t *TargetBase) GenericDesc() string {
	return t.genericDesc
}

func (t *TargetBase) String() string {
	return t.Name()
}

type Targets map[string]Target

func (ts Targets) Add(target Target) {
	n := target.Name()
	if _, ok := ts[n]; ok {
		cli.Fatalf("target %s already added", n)
	}
	_, ok := knownTargets[n]
	if !ok {
		cli.Fatalf("target %s is not known", n)
	}
	ts[n] = target
}

// KnownTargets returns a list of all allowed targets along with their generic descriptions.
func KnownTargets() map[string]TargetBase {
	return knownTargets
}

// MustGetTargetBase returns a pointer to a new copy of a known target base,
// or causes the program to fail if the named target does not exist.
func MustGetTargetBase(name string) *TargetBase {
	b, ok := knownTargets[name]
	if !ok {
		cli.Fatalf("target %s not known", name)
	}
	return &b
}

type Staler interface {
	Stale(*Context) bool
}

type DockerRunner interface {
	DockerRun(*Context) *docker.Run
}

type DockerContainer interface {
	DockerRunner
	DockerContainerName() string
}

type SetStater interface {
	SetState(string, interface{})
}

type Stater interface {
	State() interface{}
}

func (s *Sous) RunTarget(t Target, c *Context) (bool, interface{}) {
	depsRebuilt := false
	var state interface{}
	for _, d := range t.DependsOn() {
		depsRebuilt, state = s.RunTarget(d, c)
		if ss, ok := t.(SetStater); ok {
			ss.SetState(d.Name(), state)
		}
	}
	// Now we have run all dependencies, run this
	// one if necessary...
	rebuilt := s.BuildIfNecessary(t, c)
	// If this target specifies a docker run command, invoke it.
	if runner, ok := t.(DockerRunner); ok {
		run := runner.DockerRun(c)
		if run.ExitCode() != 0 {
			cli.Fatalf("Docker run failed.")
		}
	}
	// Get any available state...
	if s, ok := t.(Stater); ok {
		state = s.State()
	}
	return rebuilt || depsRebuilt, state
}

var knownTargets = map[string]TargetBase{
	"compile": TargetBase{
		name:        "compile",
		genericDesc: "a container that performs any pre-compile steps that should happen before building the app for deployment",
	},
	"app": TargetBase{
		name:        "app",
		genericDesc: "a container containing the application itself, as it would be deployed",
	},
	"test": TargetBase{
		name:        "test",
		genericDesc: "a container whose only job is to run unit tests and then exit",
	},
	"smoke": TargetBase{
		name:        "smoke",
		genericDesc: "a container whose only job is to run (smoke) tests against a remote instance of this app",
	},
}

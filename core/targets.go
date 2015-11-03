package core

import (
	"fmt"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
)

// Target describes a buildable Docker file that performs a particular task related to building
// testing and deploying the application. Each pack under packs/ will customise its targets for
// the specific jobs that need to be performed for that pack.
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
	// current project. If not, it should return an error.
	Check() error
	// Dockerfile is the shebang method which writes out a functionally complete *docker.Dockerfile
	// This method is only invoked only once the Detect func has successfully detected target availability.
	Dockerfile() *docker.Dockerfile
}

// ContainerTarget is a specialisation of Target that in addition to building a Dockerfile,
// also returns a Docker run command that can be invoked on images built from that Dockerfile, which
// the build process invokes to create a Docker container when needed.
type ContainerTarget interface {
	Target
	// DockerRun returns a Docker run command which the build process can use to
	// create the container.
	DockerRun(*Context) *docker.Run
	// ContainerName returns the name to be given to the container built by
	// this target.
	ContainerName(*Context) string
	// ContainerIsStale should return true if the container needs to be rebuilt,
	// otherwise it returns false. Certain conditions (like Sous itself being upgraded always cause root
	// and branch rebuilds, regardless of this return value.
	ContainerIsStale(*Context) bool
}

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

type ImageIsStaler interface {
	ImageIsStale(*Context) bool
}

type PreDockerBuilder interface {
	PreDockerBuild(*Context)
}

type SetStater interface {
	SetState(string, interface{})
}

type Stater interface {
	State(*Context) interface{}
}

func (s *Sous) RunTarget(t Target, c *Context) (bool, interface{}) {
	fmt.Sprintf("Running target %s", t.Name())
	depsRebuilt := false
	var state interface{}
	deps := t.DependsOn()
	if len(deps) != 0 {
		for _, d := range deps {
			cli.Logf(" ===> Building dependency %s", d.Name())
			dt, dc := s.AssembleTargetContext(d.Name())
			depsRebuilt, state = s.RunTarget(dt, dc)
			if ss, ok := t.(SetStater); ok {
				ss.SetState(dt.Name(), state)
			}
		}
		cli.Logf(" ===> All dependencies of %s built", t.Name())
	}
	// Now we have run all dependencies, run this
	// one if necessary...
	rebuilt := s.BuildImageIfNecessary(t, c)
	// If this target specifies a docker container, invoke it.
	if ct, ok := t.(ContainerTarget); ok {
		fmt.Sprintf(" ===> Running target image %s", t.Name())
		run, isNew := s.RunContainerTarget(ct, c)
		if isNew {
			cli.Logf(" ===> Preparing %s container for first use", t.Name())
		}
		if run.ExitCode() != 0 {
			cli.Fatalf("Docker run failed.")
		}
	}
	// Get any available state...
	if s, ok := t.(Stater); ok {
		state = s.State(c)
	}
	return rebuilt || depsRebuilt, state
}

func (s *Sous) RunContainerTarget(t ContainerTarget, c *Context) (*docker.Run, bool) {
	container := docker.ContainerWithName(t.ContainerName(c))
	if !container.Exists() || t.ContainerIsStale(c) || s.OverrideContainerRebuild(t, container) {
		cli.Logf("Re-using build container %s", container)
		return docker.NewReRun(container), false
	}
	if err := container.Remove(); err != nil {
		cli.Fatalf("Unable to remove outdated container %s", container)
	}
	return t.DockerRun(c), true
}

func (s *Sous) OverrideContainerRebuild(t ContainerTarget, container docker.Container) bool {
	image := container.Image()
	baseImage := t.Dockerfile().From
	return docker.BaseImageUpdated(baseImage, image)

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

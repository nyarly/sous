package core

import "github.com/opentable/sous/tools/docker"

// Target describes a buildable Docker image that performs a particular task related to building
// testing and deploying the application. Each pack under packs/ will customise its targets for
// the specific jobs that need to be performed for that pack.
type Target struct {
	// Name of the target, as used in command-line operations.
	Name,
	// GenericDesc is a generic description of the target, applicable to any pack. It is
	// used mostly for help and exposition.
	GenericDesc,
	// PackDesc is a static description of what this target does exactly in the context
	// of the pack that owns it. It should be set by the pack when it is initialised.
	PackDesc string
	// Applicable indicates whether or not this target is applicable for the
	// pack that owns it at all. If not, no attempt will be made to detect this
	// feature for that pack.
	Applicable bool
	// Detect is a function which tries to detect if this target is possible with the
	// current project. If so, it returns a complete *AppInfo, else nil and an error.
	Detect func(c *Context, packInfo interface{}) (*AppInfo, error)
	// MakeDockerfile is the shebang method which writes out a functionally complete *docker.Dockerfile
	// This method is only invoked only once the Detect func has successfully detected target availability.
	// The *AppInfo from Detect is passed in as context.
	MakeDockerfile func(a *AppInfo, packInfo interface{}) *docker.Dockerfile
	// DependsOn lists the direct dependencies of this target. Dependencies listed as "optional" will
	// always be built when available, but if they are not available will be ignored. It is the job
	// of each package under packs/ to correctly define these relationships.
	DependsOn []Target
}

// Targets contains an instance of each possible target.
type Targets struct {
	// See the text in GenericDesc of the *Targets literal defined in NewTargets below for
	// descriptions of each of these targets.
	Compile, App, Test, Smoke Target
}

// NewTargets creates a bare-bones *Targets with names and generic descriptions. All instances
// of targets should be initialised this way to ensure help and error messages remain consistent.
func NewTargets() *Targets {
	return &Targets{
		Compile: Target{
			Name:        "compile",
			GenericDesc: "a container that performs any pre-compile steps that should happen before building the app for deployment",
		},
		App: Target{
			Name:        "app",
			GenericDesc: "a container containing the application itself, as it would be deployed",
		},
		Test: Target{
			Name:        "test",
			GenericDesc: "a container whose only job is to run unit tests and then exit",
		},
		Smoke: Target{
			Name:        "smoke",
			GenericDesc: "a container whose only job is to run (smoke) tests against a remote instance of this app",
		},
	}
}

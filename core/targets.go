package core

import (
	"fmt"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/path"
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
	// Pack is the pack this Target belongs to.
	Pack() Pack
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
	ContainerIsStale(*Context) (bool, string)
}

type TargetBase struct {
	name,
	genericDesc string
	pack Pack
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

func (t *TargetBase) Pack() Pack {
	return t.pack
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
func MustGetTargetBase(name string, pack Pack) *TargetBase {
	b, ok := knownTargets[name]
	if !ok {
		cli.Fatalf("target %s not known", name)
	}
	targetCopy := b
	targetCopy.pack = pack
	return &targetCopy
}

type ImageIsStaler interface {
	ImageIsStale(*Context) (bool, string)
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

// RunTarget is used to run the top-level target from build commands.
func (s *Sous) RunTarget(t Target, c *Context) (bool, interface{}) {
	if !c.ChangesSinceLastBuild().Any() {
		if !s.Flags.ForceRebuild {
			cli.Logf("No changes since last build.")
			cli.Logf("TIP: use -rebuild to rebuild anyway, or -rebuild-all to rebuild all dependencies")
			return false, nil
		}
	}
	cli.Logf(`** ===> Building top-level target "%s"**`, t.Name())
	return s.runTarget(t, c, false)
}

func (s *Sous) runTarget(t Target, c *Context, asDependency bool) (bool, interface{}) {
	depsRebuilt := false
	var state interface{}
	deps := t.DependsOn()
	if len(deps) != 0 {
		for _, d := range deps {
			cli.Logf("** ===> Building dependency \"%s\"**", d.Name())
			dt, dc := s.AssembleTargetContext(d.Name())
			depsRebuilt, state = s.runTarget(dt, dc, true)
			if ss, ok := t.(SetStater); ok {
				ss.SetState(dt.Name(), state)
			}
		}
		cli.Logf("** ===> All dependencies of %s built**", t.Name())
	}
	// Now we have run all dependencies, run this
	// one if necessary...
	rebuilt := s.buildImageIfNecessary(t, c, asDependency)
	// If this target specifies a docker container, invoke it.
	if ct, ok := t.(ContainerTarget); ok {
		//cli.Logf("** ===> Running target image \"%s\"**", t.Name())
		run, _ := s.RunContainerTarget(ct, c, rebuilt)
		if run.ExitCode() != 0 {
			cli.Fatalf("** =x=> Docker run failed.**")
		}
	}
	// Get any available state...
	if s, ok := t.(Stater); ok {
		state = s.State(c)
	}
	return rebuilt || depsRebuilt, state
}

func (s *Sous) RunContainerTarget(t ContainerTarget, c *Context, imageRebuilt bool) (*docker.Run, bool) {
	stale, reason, container := s.NewContainerNeeded(t, c, imageRebuilt)
	if stale {
		cli.Logf("** ===> Creating new %s container because %s**", t.Name(), reason)
		if container != nil {
			if err := container.Remove(); err != nil {
				cli.Fatalf("Unable to remove outdated container %s", container)
			}
		}
		return t.DockerRun(c), true
	}
	cli.Logf("** ===> Re-using build container %s**", container)
	return docker.NewReRun(container), false
}

func (s *Sous) NewContainerNeeded(t ContainerTarget, c *Context, imageRebuilt bool) (bool, string, docker.Container) {
	container := docker.ContainerWithName(t.ContainerName(c))
	if !container.Exists() {
		container = nil
	}
	if stale, reason := t.ContainerIsStale(c); stale {
		return true, reason, container
	}
	if container == nil {
		return true, fmt.Sprintf("no container named %s exists", container), nil
	}
	if imageRebuilt {
		return true, "its underlying image was rebuilt", container
	}
	if stale, reason := s.OverrideContainerRebuild(t, container); stale {
		return true, reason, container
	}
	return false, "", container
}

func (s *Sous) OverrideContainerRebuild(t ContainerTarget, container docker.Container) (bool, string) {
	image := container.Image()
	baseImage := t.Dockerfile().From
	if docker.BaseImageUpdated(baseImage, image) {
		return true, fmt.Sprintf("base image %s updated", baseImage)
	}
	return false, ""
}

// BuildIfNecessary usually rebuilds any target if anything of the following
// are true:
//
// - No build is available at all
// - Any files in the working tree have changed
// - Sous has been updated
// - Sous config has changed
//
// However, you may override this behaviour for a specific target by implementing
// the Staler interface: { Stale(*Context) bool }
func (s *Sous) BuildImageIfNecessary(t Target, c *Context) bool {
	return s.buildImageIfNecessary(t, c, false)
}

func (s *Sous) buildImageIfNecessary(t Target, c *Context, asDependency bool) bool {
	stale, reason := s.NeedsToBuildNewImage(t, c, asDependency)
	if !stale {
		return false
	}
	cli.Logf("** ===> Rebuilding image for %s because %s**", t.Name(), reason)
	s.BuildImage(t, c)
	return true
}

// NeedsBuild detects if the project's last
// build is stale, and if it therefore needs to be rebuilt. This can be overidden
// by implementing the Staler interfact on individual build targets. This default
// implementation rebuilds on absolutely any change in sous (i.e. new version/new
// config) or in the working tree (new or modified files).
func (s *Sous) NeedsToBuildNewImage(t Target, c *Context, asDependency bool) (bool, string) {
	if s.Flags.ForceRebuildAll {
		return true, "-rebuild-all flag was used"
	}
	if s.Flags.ForceRebuild && !asDependency {
		return true, "-rebuild flag was used"
	}
	changes := c.ChangesSinceLastBuild()
	if staler, ok := t.(ImageIsStaler); ok {
		if stale, reason := staler.ImageIsStale(c); stale {
			return true, reason
		}
	} else if changes.Any() {
		reason := "changes were detected"
		if changes.WorkingTreeChanged {
			reason = "your working tree has changed"
		} else if changes.NewCommit {
			reason = "you have a new commit"
		} else if changes.NoBuiltImage {
			reason = "no corresponding image exists yet"
		} else if changes.SousUpdated {
			reason = "sous itself was updated"
		}
		return true, reason
	}
	// Always force a rebuild if is base image has been updated.
	baseImage := t.Dockerfile().From
	// TODO: This is probably a bit too aggressive, consider only asking the user to
	// update base images every 24 hours, if they have actually been updated.
	s.UpdateBaseImage(baseImage)
	if c.BuildNumber() == 1 {
		return true, fmt.Sprintf("there are no successful builds yet for the current revision (%s)", c.Git.CommitSHA)
	}
	if !c.LastBuildImageExists() {
		return true, fmt.Sprintf("the last successful build image no longer exists (%s)", c.PrevDockerTag())
	}
	if docker.BaseImageUpdated(baseImage, c.PrevDockerTag()) {
		return true, fmt.Sprintf("the base image %s was updated", baseImage)
	}
	// Always force a build if Sous itself has been updated
	// NB: Always keep this check until last, since it's annoying, so only report this as the reason to rebuild
	// if none of the reasons above hold true. Sous does its own PR innit.
	if changes.SousUpdated {
		return true, fmt.Sprintf("sous itself or its config was updated")
	}
	return false, ""
}

func (s *Sous) BuildImage(t Target, c *Context) {
	c.IncrementBuildNumber()
	if file.Exists("Dockerfile") {
		cli.Logf("**WARNING: Your local Dockerfile is ignored by sous, use `sous dockerfile %s` to see the dockerfile being used here**", t.Name())
	}
	dfPath := path.Resolve(c.FilePath("Dockerfile"))
	if prebuilder, ok := t.(PreDockerBuilder); ok {
		prebuilder.PreDockerBuild(c)
	}
	// NB: Always rebuild the Dockerfile after running pre-build, since pre-build
	// may update target state to reflect things like copied file locations etc.
	s.WriteDockerfile(t, c)
	docker.BuildFile(dfPath, ".", c.DockerTag())
	c.Commit()
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

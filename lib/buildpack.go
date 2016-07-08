package sous

import "time"

type (
	// Builder defines a container-based build system.
	Builder interface {
		Registry // TODO: Remove registry from here.
		// Build performs a build and returns the result.
		Build(*BuildContext, Buildpack, *DetectResult) (*BuildResult, error)
	}
	BuildArtifact struct {
		Name, Type string
	}
	// Buildpack is a set of instructions used to build a particular
	// kind of project.
	Buildpack interface {
		Detect(*BuildContext) (*DetectResult, error)
		Build(*BuildContext) (*BuildResult, error)
	}
	// DetectResult represents the result of a detection.
	DetectResult struct {
		Compatible  bool
		Description string
		Data        interface{}
	}
	// BuildResult represents the result of a build made with a Buildpack.
	BuildResult struct {
		ImageID                   string
		VersionName, RevisionName string
		Elapsed                   time.Duration
	}
)

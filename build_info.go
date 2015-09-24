package main

import "fmt"

type BuildInfo struct {
	Context        *BuildContext
	Version        string
	DockerRegistry string
}

func (b *BuildInfo) DockerImage() string {
	// e.g. docker.otenv.com/widget-factory:v0.12.1-ci-912eeeab-1
	return fmt.Sprintf("%s/%s:v%s-ci-%s-%d",
		b.DockerRegistry,
		b.Context.CanonicalPackageName(),
		b.Version,
		b.Context.Git.CommitSHA[0:8],
		b.Context.BuildNumber)
}

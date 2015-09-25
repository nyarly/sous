package build

import "fmt"

type BuildInfo struct {
	Context *BuildContext
	App     *AppInfo
}

func (b *BuildInfo) DockerImage() string {
	// e.g. on TeamCity:
	//   docker.otenv.com/widget-factory:v0.12.1-ci-912eeeab-1
	if b.Context.IsCI() {
		return fmt.Sprintf("%s/%s:v%s-ci-%s-%d",
			b.Context.DockerRegistry,
			b.Context.CanonicalPackageName(),
			b.App.Version,
			b.Context.Git.CommitSHA[0:8],
			b.Context.BuildNumber)
	}
	// e.g. on local dev machine:
	//   docker.otenv.com/widget-factory:username@host-v0.12.1-912eeeab-1
	return fmt.Sprintf("%s/%s/%s:v%s-%s-%s-%d",
		b.Context.DockerRegistry,
		b.Context.User,
		b.Context.CanonicalPackageName(),
		b.App.Version,
		b.Context.Git.CommitSHA[0:8],
		b.Context.Host,
		b.Context.BuildNumber)
}

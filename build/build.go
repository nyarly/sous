package build

import . "github.com/opentable/sous/util"

func Build(args []string) {
	gitVersion := Cmd("git", "version")
	Logf(gitVersion)
	context := getBuildContext()
	info := &BuildInfo{Context: context}
	info.App = tryBuildNodeJS(context)
	successOnAppInfo(info)
	Dief("no buildable project detected")
}

func BuildHelp() string {
	return `sous build detects your project type, and tries to find a matching
OpenTable supported stack to build against. Right now it only supports NodeJS
projects. It builds a docker image, tagged and labelled correctly.

sous build does not have any options yet`
}

func successOnAppInfo(i *BuildInfo) {
	if i.App == nil {
		return
	}
	d := i.App.Dockerfile
	prefix := "com.opentable"
	d.AddLabel(prefix+".builder.app", "sous")
	d.AddLabel(prefix+".builder.host", i.Context.Host)
	d.AddLabel(prefix+".builder.fullhost", i.Context.FullHost)
	d.AddLabel(prefix+".builder.user", i.Context.User)
	d.AddLabel(prefix+".source.git.repo", i.Context.Git.CanonicalName())
	d.AddLabel(prefix+".source.git.commit-sha", i.Context.Git.CommitSHA)

	d.Maintainer = i.Context.User

	WriteFile(i.App.Dockerfile.Render(), "Dockerfile")

	ExitSuccessf("Successfully built %s v%s as %s",
		i.Context.CanonicalPackageName(),
		i.App.Version,
		i.DockerImage())
}

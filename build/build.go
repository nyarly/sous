package build

import . "github.com/opentable/sous/util"

func Build() {
	gitVersion := Cmd("git", "version")
	Logf(gitVersion)
	context := getBuildContext()
	info := &BuildInfo{Context: context}
	info.App = tryBuildNodeJS(context)
	successOnAppInfo(info)
	Dief("no buildable project detected")
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

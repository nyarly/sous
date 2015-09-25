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
	d.AddLabel("com.opentable.builder.app", "sous")
	d.AddLabel("com.opentable.source.git.repo", i.Context.Git.CanonicalName())
	d.AddLabel("com.opentable.source.git.commit-sha", i.Context.Git.CommitSHA)

	WriteFile(i.App.Dockerfile.Render(), "Dockerfile")

	ExitSuccessf("Successfully built %s v%s as %s",
		i.Context.CanonicalPackageName(),
		i.App.Version,
		i.DockerImage())
}

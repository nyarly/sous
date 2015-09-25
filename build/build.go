package build

import . "github.com/opentable/sous/util"

func Build() {
	// Ensure dependencies are installed:
	//    - Git
	// Sniff current directory for type:
	//
	//    - NodeJS = package.json
	//    - others later
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
	ExitSuccessf("Successfully built %s v%s as %s",
		i.Context.CanonicalPackageName(),
		i.App.Version,
		i.DockerImage())
}

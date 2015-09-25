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
	buildContext := getBuildContext()
	var info *BuildInfo
	info = tryBuildNodeJS(buildContext)
	successOnInfo(info)
	Dief("no buildable project detected")
}

func successOnInfo(i *BuildInfo) {
	ExitSuccessf("Successfully built %s v%s as %s",
		i.Context.CanonicalPackageName(),
		i.Version,
		i.DockerImage())
}

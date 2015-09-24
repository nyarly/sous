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
	tryBuildNodeJS(buildContext)
	Dief("no buildable project detected")
}

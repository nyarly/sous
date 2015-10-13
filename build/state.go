package build

import (
	"fmt"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/dir"
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/git"
)

type BuildState struct {
	CommitSHA, LastCommitSHA string
	Commits                  map[string]*Commit
	path                     string
}

func GetBuildState(action string, g *git.Info) *BuildState {
	filePath := getStateFile(action, g)
	var state *BuildState
	if !file.ReadJSON(&state, filePath) {
		state = &BuildState{
			Commits: map[string]*Commit{},
		}
	}
	if state == nil {
		cli.Fatalf("Nil state at %s", filePath)
	}
	if state.Commits == nil {
		cli.Fatalf("Nil commits at %s", filePath)
	}
	c, ok := state.Commits[g.CommitSHA]
	if !ok {
		state.Commits[g.CommitSHA] = &Commit{}
	}
	state.LastCommitSHA = state.CommitSHA
	state.CommitSHA = g.CommitSHA
	state.path = filePath
	c = state.Commits[g.CommitSHA]
	if buildingInCI() {
		bn, ok := tryGetBuildNumberFromEnv()
		if !ok {
			cli.Fatalf("unable to get build number from $BUILD_NUMBER TeamCity")
		}
		c.BuildNumber = bn
	}
	c.OldHash = c.Hash
	c.Hash = CalculateHash()
	return state
}

func getStateFile(action string, g *git.Info) string {
	dirPath := fmt.Sprintf("~/.sous/builds/%s/%s", g.CanonicalName(), action)
	dir.EnsureExists(dirPath)
	return fmt.Sprintf("%s/state", dirPath)
}

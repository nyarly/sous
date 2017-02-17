package config

import (
	"testing"

	"github.com/nyarly/testify/assert"
	"github.com/opentable/sous/ext/github"
	sous "github.com/opentable/sous/lib"
)

func TestDeployFilter(t *testing.T) {
	shc := sous.SourceHostChooser{SourceHosts: []sous.SourceHost{github.SourceHost{}}}

	dep := func(repo, offset, flavor string) *sous.Deployment {
		return &sous.Deployment{
			SourceID: sous.SourceID{
				Location: sous.SourceLocation{
					Repo: repo,
					Dir:  offset,
				},
			},
			Flavor: flavor,
		}
	}

	deploys := []*sous.Deployment{
		dep("github.com/opentable/example", "", ""),
		dep("github.com/opentable/other", "", ""),
		dep("github.com/opentable/example", "somewhere", ""),
		dep("github.com/opentable/flavored", "", "choc"),
		dep("github.com/opentable/flavored", "", "van"),
	}

	testFilter := func(df DeployFilterFlags, idxs ...int) {
		rf, err := df.BuildFilter(shc.ParseSourceLocation)
		assert.NoError(t, err)

		for n, dep := range deploys {
			if len(idxs) > 0 && idxs[0] == n {
				assert.True(t, rf.FilterDeployment(dep), "%v doesn't match #%d %v", rf, n, dep)
				idxs = idxs[1:]
			} else {
				assert.False(t, rf.FilterDeployment(dep), "%v matches #%d %v", rf, n, dep)
			}
		}
	}

	testFilter(DeployFilterFlags{All: true}, 0, 1, 2, 3, 4)
	testFilter(DeployFilterFlags{Repo: deploys[0].SourceID.Location.Repo}, 0)
	testFilter(DeployFilterFlags{Repo: deploys[1].SourceID.Location.Repo}, 1)
	testFilter(DeployFilterFlags{}, 0, 1)
	testFilter(DeployFilterFlags{Offset: ""}, 0, 1)
	testFilter(DeployFilterFlags{Offset: "*"}, 0, 1, 2)
	testFilter(DeployFilterFlags{Offset: "*", Flavor: "*"}, 0, 1, 2, 3, 4)
	testFilter(DeployFilterFlags{Flavor: "choc"}, 3)

}

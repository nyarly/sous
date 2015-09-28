package nodejs

import (
	"fmt"
	"strings"

	. "github.com/opentable/sous/tools"
	"github.com/opentable/sous/tools/docker"
	"github.com/wmark/semver"
)

type NodePackage struct {
	Name    string
	Version string
	Engines NodePackageEngines
	Scripts NodePackageScripts
}
type NodePackageEngines struct {
	Node, NPM string
}
type NodePackageScripts struct {
	Start, Test string
}

var availableNodeVersions = []string{
	"0.12.7",
	"0.10.40",
}

func mustVersion(v *semver.Version, err error) *semver.Version {
	if err != nil {
		Dief("Unable to parse version; %s", err)
	}
	return v
}

func buildNodeJS(np *NodePackage) *docker.Dockerfile {
	var nodeVersion string
	if np.Engines.Node == "" {
		nodeVersion = availableNodeVersions[0]
		Logf("WARNING: No NodeJS version specified in package.json; using latest available version (%s)", nodeVersion)
	} else {
		nodeVersion = selectBestVersion(np.Engines.Node, availableNodeVersions)
		if nodeVersion == "" {
			Dief("unable to satisfy NodeJS version '%s' (from package.json); available versions are: %s", np.Engines.Node, strings.Join(availableNodeVersions, ", "))
		}
	}
	baseImageTag := "latest"
	from := fmt.Sprintf("docker.otenv.com/ot-node-base-%s:%s", nodeVersion, baseImageTag)
	df := &docker.Dockerfile{
		From:    from,
		Add:     []docker.Add{docker.Add{Files: []string{"."}, Dest: "/srv/app"}},
		Workdir: "/srv/app",
		Run:     []string{"npm install --production; ls -la /srv/app"},
		CMD:     Whitespace.Split(np.Scripts.Start, -1),
	}
	df.AddLabel("com.opentable.stack", "NodeJS")
	df.AddLabel("com.opentable.stack.nodejs.version", nodeVersion)
	return df
}

func selectBestVersion(rangeSpecifier string, from []string) string {
	r, err := semver.NewRange(rangeSpecifier)
	if err != nil {
		Dief("Unable to parse version range '%s' from package.json engines directive", rangeSpecifier)
	}
	for _, vs := range from {
		v := mustVersion(semver.NewVersion(vs))
		if r.IsSatisfiedBy(v) {
			return vs
		}
	}
	return ""
}

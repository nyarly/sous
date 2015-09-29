package nodejs

import (
	"fmt"
	"strings"

	. "github.com/opentable/sous/tools"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/version"
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

var availableNodeVersions = version.VersionList("0.12.7", "0.10.40")

func bestSupportedNodeVersion(np *NodePackage) string {
	var nodeVersion *version.V
	if np.Engines.Node == "" {
		nodeVersion = availableNodeVersions[0]
		Logf("WARNING: No NodeJS version specified in package.json; using latest available version (%s)", nodeVersion)
	} else {
		nodeVersion = version.Range(np.Engines.Node).BestMatchFrom(availableNodeVersions)
		if nodeVersion == nil {
			Dief("unable to satisfy NodeJS version '%s' (from package.json); available versions are: %s",
				np.Engines.Node, strings.Join(availableNodeVersions.Strings(), ", "))
		}
	}
	return nodeVersion.String()
}

func dockerFrom(np *NodePackage, nodeVersion string) string {
	baseImageTag := "latest"
	return fmt.Sprintf("docker.otenv.com/ot-node-base-%s:%s", nodeVersion, baseImageTag)
}

func buildNodeJS(np *NodePackage) *docker.Dockerfile {
	nodeVersion := bestSupportedNodeVersion(np)
	from := dockerFrom(np, nodeVersion)
	df := &docker.Dockerfile{
		From:    from,
		Add:     []docker.Add{docker.Add{Files: []string{"."}, Dest: "/srv/app"}},
		Workdir: "/srv/app",
		Run:     []string{"npm install --registry=http://artifactory.otenv.com/artifactory/api/npm/npm-virtual --production; ls -la /srv/app"},
		CMD:     Whitespace.Split(np.Scripts.Start, -1),
	}
	df.AddLabel("com.opentable.stack", "NodeJS")
	df.AddLabel("com.opentable.stack.nodejs.version", nodeVersion)
	return df
}

func testNodeJS(np *NodePackage) *docker.Dockerfile {
	nodeVersion := bestSupportedNodeVersion(np)
	from := dockerFrom(np, nodeVersion)
	df := &docker.Dockerfile{
		From:    from,
		Add:     []docker.Add{docker.Add{Files: []string{"."}, Dest: "/srv/app"}},
		Workdir: "/srv/app",
		Run:     []string{"npm install --registry=http://artifactory.otenv.com/artifactory/api/npm/npm-virtual; ls -la /srv/app"},
		CMD:     []string{"/usr/local/bin/npm", "test"},
	}
	df.AddLabel("com.opentable.stack", "NodeJS")
	df.AddLabel("com.opentable.stack.nodejs.version", nodeVersion)
	df.AddLabel("com.opentable.tests", "true")
	return df
}

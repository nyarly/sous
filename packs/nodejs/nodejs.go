package nodejs

import (
	"fmt"
	"strings"

	"github.com/opentable/sous/tools"
	"github.com/opentable/sous/tools/cli"
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

var availableNodeVersions = version.VersionList("4.1.0", "4.0.0", "0.12.7", "0.10.40")

func bestSupportedNodeVersion(np *NodePackage) string {
	var nodeVersion *version.V
	if np.Engines.Node == "" {
		nodeVersion = availableNodeVersions[0]
		cli.Logf("WARNING: No NodeJS version specified in package.json; using latest available version (%s)", nodeVersion)
	} else {
		nodeVersion = version.Range(np.Engines.Node).BestMatchFrom(availableNodeVersions)
		if nodeVersion == nil {
			cli.Fatalf("unable to satisfy NodeJS version '%s' (from package.json); available versions are: %s",
				np.Engines.Node, strings.Join(availableNodeVersions.Strings(), ", "))
		}
	}
	return nodeVersion.String()
}

func dockerFrom(np *NodePackage, nodeVersion string) string {
	baseImageTag := "latest"
	return fmt.Sprintf("docker.otenv.com/ot-node-base-%s:%s", nodeVersion, baseImageTag)
}

var npmVersions = version.VersionList("3.3.4", "2.4.15")
var defaultNPMVersion = version.Version("2.4.15")
var npmRegistry = "http://artifactory.otenv.com/artifactory/api/npm/npm-virtual"

func baseDockerfile(np *NodePackage) *docker.Dockerfile {
	nodeVersion := bestSupportedNodeVersion(np)
	from := dockerFrom(np, nodeVersion)
	npmVer := defaultNPMVersion
	if np.Engines.NPM != "" {
		npmVer = version.Range(np.Engines.NPM).BestMatchFrom(npmVersions)
	}
	df := &docker.Dockerfile{
		From:    from,
		Add:     []docker.Add{docker.Add{Files: []string{"."}, Dest: "/srv/app"}},
		Workdir: "/srv/app",
	}
	df.AddRun("npm install npm@%s", npmVer)
	df.AddLabel("com.opentable.stack", "NodeJS")
	df.AddLabel("com.opentable.stack.nodejs.version", nodeVersion)
	return df
}

func buildNodeJS(np *NodePackage) *docker.Dockerfile {
	df := baseDockerfile(np)
	// Pick out the contents of NPM start to invoke directly (using npm start in
	// production shields the app from signals, which are required to be handled by
	// the app itself to do graceful shutdown.
	df.CMD = tools.Whitespace.Split(np.Scripts.Start, -1)
	df.AddRun("npm install --registry=%s --production", npmRegistry)
	return df
}

func testNodeJS(np *NodePackage) *docker.Dockerfile {
	df := baseDockerfile(np)
	df.CMD = []string{"npm", "test"}
	df.AddRun("npm install --registry=%s", npmRegistry)
	df.AddLabel("com.opentable.tests", "true")
	return df
}

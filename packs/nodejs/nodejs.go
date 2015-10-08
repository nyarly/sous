package nodejs

import (
	"strings"

	"github.com/opentable/sous/config"
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

func keys(m map[string]string) []string {
	ks := make([]string, len(m))
	i := 0
	for k := range m {
		ks[i] = k
		i++
	}
	return ks
}

func bestSupportedNodeVersion(np *NodePackage) string {
	var nodeVersion *version.V
	c := config.Load()
	availableNodeVersions := version.VersionList(strings.Join(keys(c.Packs.NodeJS.NodeVersionsToDockerBaseImages), ","))
	nodeVersion = version.Range(np.Engines.Node).BestMatchFrom(availableNodeVersions)
	if nodeVersion == nil {
		cli.Fatalf("unable to satisfy NodeJS version '%s' (from package.json); available versions are: %s",
			np.Engines.Node, strings.Join(availableNodeVersions.Strings(), ", "))
	}
	return nodeVersion.String()
}

func dockerFrom(np *NodePackage, nodeVersion string) string {
	c := config.Load()
	return c.Packs.NodeJS.NodeVersionsToDockerBaseImages[nodeVersion]
}

var npmVersions = version.VersionList("3.3.4", "2.4.15")
var defaultNPMVersion = version.Version("2.4.15")
var npmRegistry = config.Load().Packs.NodeJS.NPMMirrorURL

var wd = "/srv/app/"

func baseDockerfile(np *NodePackage) *docker.Dockerfile {
	nodeVersion := bestSupportedNodeVersion(np)
	from := dockerFrom(np, nodeVersion)
	npmVer := defaultNPMVersion
	if np.Engines.NPM != "" {
		npmVer = version.Range(np.Engines.NPM).BestMatchFrom(npmVersions)
	}
	df := &docker.Dockerfile{
		From:    from,
		Add:     []docker.Add{docker.Add{Files: []string{"."}, Dest: wd}},
		Workdir: wd,
	}
	npmMajorVer := npmVer.String()[0:1]
	df.AddRun("npm install -g npm@%s", npmMajorVer)
	df.AddLabel("com.opentable.stack", "NodeJS")
	df.AddLabel("com.opentable.stack.nodejs.version", nodeVersion)
	return df
}

func buildNodeJS(np *NodePackage) *docker.Dockerfile {
	df := baseDockerfile(np)
	// Pick out the contents of NPM start to invoke directly (using npm start in
	// production shields the app from signals, which are required to be handled by
	// the app itself to do graceful shutdown.
	df.AddRun("npm install --registry=%s --production", npmRegistry)
	df.CMD = tools.Whitespace.Split(np.Scripts.Start, -1)
	return df
}

func testNodeJS(np *NodePackage) *docker.Dockerfile {
	df := baseDockerfile(np)
	df.AddRun("cd "+wd+" && npm install --registry=%s", npmRegistry)
	df.AddLabel("com.opentable.tests", "true")
	df.CMD = []string{"npm", "test"}
	return df
}

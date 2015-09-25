package build

import (
	"fmt"
	"strings"

	. "github.com/opentable/sous/util"
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

func tryBuildNodeJS(bc *BuildContext) *AppInfo {
	var np *NodePackage
	if !ReadFileJSON(&np, "package.json") {
		return nil
	}
	return buildNodeJS(bc, np)
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

func buildNodeJS(bc *BuildContext, np *NodePackage) *AppInfo {
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
	df := &Dockerfile{
		From:    from,
		Add:     []Add{Add{Files: []string{"."}, Dest: "/srv/app"}},
		Workdir: "/srv/app",
		Run:     []string{"npm install --production; ls -la /srv/app"},
		CMD:     Whitespace.Split(np.Scripts.Start, -1),
	}
	return &AppInfo{
		Version:    np.Version,
		Dockerfile: df,
	}
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

package nodejs

import (
	"strings"

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
	Start, Test, InstallProduction string
}

func (p *Pack) AvailableNodeVersions() version.VL {
	vs := make([]*version.V, len(*p.Config.AvailableVersions))
	for i, v := range *p.Config.AvailableVersions {
		vs[i] = version.Version(v.Name)
	}
	return vs
}

func (p *Pack) bestSupportedNodeVersion() string {
	np := p.PackageJSON
	var nodeVersion *version.V
	nodeVersion = version.Range(np.Engines.Node).BestMatchFrom(p.AvailableNodeVersions())
	if nodeVersion == nil {
		cli.Fatalf("unable to satisfy NodeJS version '%s' (from package.json); available versions are: %s",
			np.Engines.Node, strings.Join(p.AvailableNodeVersions().Strings(), ", "))
	}
	return nodeVersion.String()
}

func (p *Pack) dockerFrom(nodeVersion, target string) string {
	if tag, ok := p.Config.AvailableVersions.GetBaseImageTag(nodeVersion, target); ok {
		return tag
	}
	cli.Fatalf("No base image available for NodeJS %s, target: %s", nodeVersion, target)
	return ""
}

var npmVersions = version.VersionList("3.3.4", "2.4.15")
var defaultNPMVersion = version.Version("2.4.15")

var wd = "/srv/app/"

func (p *Pack) baseDockerfile(target string) *docker.Dockerfile {
	np := p.PackageJSON
	nodeVersion := p.bestSupportedNodeVersion()
	from := p.dockerFrom(nodeVersion, target) + ":latest"
	npmVer := defaultNPMVersion
	if np.Engines.NPM != "" {
		npmVer = version.Range(np.Engines.NPM).BestMatchFrom(npmVersions)
		if npmVer == nil {
			cli.Logf("NPM version %s not supported, try using a range instead.", np.Engines.NPM)
			cli.Fatalf("Available NPM version ranges are: '^3' and '^2'")
		}
	}
	df := &docker.Dockerfile{
		From:        from,
		Add:         []docker.Add{docker.Add{Files: []string{"."}, Dest: wd}},
		Workdir:     wd,
		LabelPrefix: "com.opentable",
	}
	df.AddLabel("build.pack.nodejs.version", nodeVersion)
	return df
}

package build

import "github.com/opentable/sous/tools/docker"

type Pack struct {
	Name                  string
	Detect                func() error
	Features              map[string]*Feature
	CompatibleProjectDesc func() string
	CheckCompatibility    func() []string
}

type Feature struct {
	Detect         func(*Context) (*AppInfo, error)
	MakeDockerfile func(*AppInfo) *docker.Dockerfile
}

package build

import "github.com/opentable/sous/tools/docker"

type Pack struct {
	Name     string
	Detect   func() error
	Features map[string]*Feature
}

type Feature struct {
	Detect         func(*Context) (*AppInfo, error)
	MakeDockerfile func(*AppInfo) *docker.Dockerfile
}

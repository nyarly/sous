package core

import "github.com/opentable/sous/tools/docker"

type Target struct {
	Detect         func(*Context) (*AppInfo, error)
	MakeDockerfile func(*AppInfo) *docker.Dockerfile
}

package core

import "github.com/opentable/sous/tools/docker"

type Target struct {
	Detect         func(c *Context, packInfo interface{}) (*AppInfo, error)
	MakeDockerfile func(a *AppInfo, packInfo interface{}) *docker.Dockerfile
}

package core

import (
	"fmt"

	"github.com/opentable/sous/tools/version"
)

type Config struct {
	DockerRegistry    string
	DockerLabelPrefix string
	GlobalDockerTags  Values
	Packs             *Packs
	Platform          *Platform
	// ContractDefs maps a service kind to an ordered set of contracts
	// to run against apps of that kind.
	ContractDefs map[string]List
}

type Platform struct {
	Services []Service
	EnvDef   []EnvVar
	Envs     []Env
}

type EnvVar struct {
	Name, Type          string
	Required, Protected bool
}

// Service defines a common platform service that most apps will
// rely on. Examples include discovery servers, proxies, config servers, etc.
// These are used in local development, and may be referred to by their name
// in contracts.
type Service struct {
	Name, DockerImage, DockerRunOpts string
}

// Environment defines a named execution environment. This is an open
// ended concept, but a common usage is to have a single environment
// per datacentre, for example.
type Env struct {
	Name string
	Vars map[string]string
}

type Packs struct {
	NodeJS *NodeJSConfig
	Go     *GoConfig
}

type NodeJSConfig struct {
	AvailableVersions    *StackVersions
	DockerTags           map[string]string
	AvailableNPMVersions []string
	DefaultNodeVersion   string
}

type GoConfig struct {
	AvailableVersions *StackVersions
	DefaultGoVersion  string
}

type StackVersions []*StackVersion

type StackVersion struct {
	Name, DefaultImage string
	TargetImages       BaseImageSet
}

type BaseImageSet map[string]string

func (sv StackVersion) Version() (*version.V, error) {
	return version.NewVersion(sv.Name)
}

func (svs StackVersions) GetBaseImageTag(version, target string) (string, bool) {
	for _, sv := range svs {
		if sv.Name == version {
			if specificImage, ok := sv.TargetImages[target]; ok {
				return specificImage, true
			}
			return sv.DefaultImage, true
		}
	}
	return "", false
}

// TODO: These next 3 funcs seem terribly complex for what we're trying
// to achieve. Consider changing StackVersion to use a version range
// which is parsed on deserialisation, to avoid all this error checking.
func (svs StackVersions) ConcreteVersions() ([]*version.V, error) {
	vs := make([]*version.V, len(svs))
	for i, sv := range svs {
		v, err := sv.Version()
		if err != nil {
			return nil, err
		}
		vs[i] = v
	}
	return vs, nil
}

func (svs StackVersions) Version(ver *version.V) (*StackVersion, error) {
	for _, sv := range svs {
		v, err := sv.Version()
		if err != nil {
			return nil, err
		}
		if *v == *ver {
			return sv, nil
		}
	}
	return nil, fmt.Errorf("version %q not found", ver)
}

func (svs StackVersions) GetBestStackVersion(versionRange *version.R) (*StackVersion, error) {
	versionNames, err := svs.ConcreteVersions()
	if err != nil {
		return nil, err
	}
	v := versionRange.BestMatchFrom(versionNames)
	return svs.Version(v)
}

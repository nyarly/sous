package config

type Config struct {
	DockerRegistry    string
	DockerLabelPrefix string
	GlobalDockerTags  map[string]string
	Packs             *Packs
}

type Packs struct {
	NodeJS *NodeJSConfig
}

type NodeJSConfig struct {
	AvailableVersions    *StackVersions
	DockerTags           map[string]string
	AvailableNPMVersions []string
	DefaultNodeVersion   string
}

type StackVersions []*StackVersion

type StackVersion struct {
	Name, DefaultImage string
	TargetImages       BaseImageSet
}

type BaseImageSet map[string]string

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

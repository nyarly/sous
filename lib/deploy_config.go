package sous

import "fmt"

type (
	// DeployConfig represents the configuration of a deployment's tasks,
	// in a specific cluster. i.e. their resources, environment, and the number
	// of instances.
	DeployConfig struct {
		// Resources represents the resources each instance of this software
		// will be given by the execution environment.
		Resources Resources `yaml:",omitempty" validate:"keys=nonempty,values=nonempty"`
		// Env is a list of environment variables to set for each instance of
		// of this deployment. It will be checked for conflict with the
		// definitions found in State.Defs.EnvVars, and if not in conflict
		// assumes the greatest priority.
		Args []string `yaml:",omitempty" validate:"values=nonempty"`
		Env  `yaml:",omitempty" validate:"keys=nonempty,values=nonempty"`
		// NumInstances is a guide to the number of instances that should be
		// deployed in this cluster, note that the actual number may differ due
		// to decisions made by Sous. If set to zero, Sous will decide how many
		// instances to launch.
		NumInstances int

		// Volumes lists the volume mappings for this deploy
		Volumes Volumes
	}
)

func (dc *DeployConfig) String() string {
	return fmt.Sprintf("#%d %+v : %+v %+v", dc.NumInstances, dc.Resources, dc.Env, dc.Volumes)
}

// Equal is used to compare DeployConfigs
func (dc *DeployConfig) Equal(o DeployConfig) bool {
	Log.Vomit.Printf("%+ v ?= %+ v", dc, o)
	return (dc.NumInstances == o.NumInstances && dc.Env.Equal(o.Env) && dc.Resources.Equal(o.Resources) && dc.Volumes.Equal(o.Volumes))
}

func flattenDeployConfigs(dcs []DeployConfig) DeployConfig {
	dc := DeployConfig{
		Resources: make(Resources),
		Env:       make(Env),
	}
	for _, c := range dcs {
		if c.NumInstances != 0 {
			dc.NumInstances = c.NumInstances
			break
		}
	}
	for _, c := range dcs {
		if len(c.Volumes) != 0 {
			dc.Volumes = c.Volumes
			break
		}
	}
	for _, c := range dcs {
		if len(c.Args) != 0 {
			dc.Args = c.Args
			break
		}
	}
	for _, c := range dcs {
		for n, v := range c.Resources {
			if _, set := dc.Resources[n]; !set {
				dc.Resources[n] = v
			}
		}
		for n, v := range c.Env {
			if _, set := dc.Env[n]; !set {
				dc.Env[n] = v
			}
		}
	}
	return dc
}

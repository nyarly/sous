package deploy

import (
	"fmt"

	"github.com/imdario/mergo"
)

func (s *State) Validate() error {
	// Check that none of the manifests overwrite the protected
	// env vars.
	for _, manifest := range s.Manifests {
		for _, deployment := range manifest.Deployments {
			for _, envVar := range *s.EnvironmentDefs["Universal"] {
				if _, exists := deployment.Environment[envVar.Name]; exists {
					return fmt.Errorf(
						"%s overrides protected environment variable %s",
						manifest.App.SourceRepo, envVar)
				}
			}
		}
	}
	return nil
}

func (s *State) Merge() (State, error) {
	m := *s
	for sourceRepo, source := range s.Manifests {
		mergedManifest, err := MergeManifest(source, s)
		if err != nil {
			return State{}, err
		}
		m.Manifests[sourceRepo] = mergedManifest
	}
	return m, nil
}

func MergeManifest(source Manifest, s *State) (Manifest, error) {
	merged := source
	merged.Deployments = map[string]Deployment{}
	// Copy every non-global DC to the map (we don't want to destroy the
	// original source values, and we want to explicitly not end up with a
	// datacentre named "Global").
	for k, v := range source.Deployments {
		if k == "Global" {
			continue
		}
		merged.Deployments[k] = v
	}
	// If we have a global deployment, fill in empty values in existing
	// deployments using this, and add missing deployments.
	if global, ok := source.Deployments["Global"]; ok {
		for k := range s.Datacentres {
			// If we already have a defined deploy with this name, fill
			// in the blanks from Global...
			if defined, ok := merged.Deployments[k]; ok {
				mergo.Merge(&defined, global)   // fill in empty values
				merged.Deployments[k] = defined // copy it back to the map
			} else {
				merged.Deployments[k] = global
			}
			// Now add datacentre-wide environment variables
			for ke, ve := range s.Datacentres[k].Env {
				merged.Deployments[k].Environment[ke] = ve
			}
		}
	}
	return merged, nil
}

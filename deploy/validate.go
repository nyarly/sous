package deploy

import "fmt"

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

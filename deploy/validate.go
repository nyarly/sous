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

func (c *Contract) ValidateTest() error {
	numTests := len(c.SelfTest.CheckTests)
	numChecks := len(c.Checks)
	if numTests != numChecks {
		return fmt.Errorf("contract test %q has %d check tests; want %d",
			c.SelfTest.ContractName, numTests, numChecks)
	}
	// Check that each check has a test in the right order
	for i, check := range c.Checks {
		test := c.SelfTest.CheckTests[i]
		if test.CheckName != check.Name {
			return fmt.Errorf("Contract test %q has check test %q at position %d; want %q", c.SelfTest.ContractName, test.CheckName, i, check.Name)
		}
	}
	return nil
}

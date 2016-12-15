package sous

// A Deployable is the pairing of a Deployment and the resolved image that can
// (or has) be used to deploy it.
type Deployable struct {
	*Deployment
	*BuildArtifact
}
package singularity

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/opentable/go-singularity/dtos"
	"github.com/opentable/sous/lib"
	"github.com/opentable/sous/util/coaxer"
)

// adsBuilder represents the building of a single sous.DeployStates from a
// single Singularity-hosted cluster.
type adsBuilder struct {
	Context       context.Context
	ClientFactory func(*sous.Cluster) DeployReader
	Clusters      sous.Clusters
	Registry      sous.Registry
	ErrorCallback func(error)
}

func newADSBuilder(ctx context.Context, client func(*sous.Cluster) DeployReader, reg sous.Registry, clusters sous.Clusters) *adsBuilder {
	return &adsBuilder{
		ClientFactory: client,
		Registry:      reg,
		Clusters:      clusters,
		ErrorCallback: func(err error) { log.Println(err) },
		Context:       ctx,
	}
}

// DeployStates returns all deploy states.
func (ab *adsBuilder) DeployStates() (sous.DeployStates, error) {

	log.Printf("Getting all requests...")

	promises := make(map[string]coaxer.Promise, len(ab.Clusters))

	var requests []*dtos.SingularityRequestParent

	// Grab the list of all requests from all clusters.
	for clusterName, cluster := range ab.Clusters {
		cluster := cluster
		// TODO: Make sous.Clusters a slice to avoid this double-entry record keeping.
		cluster.Name = clusterName
		promises[cluster.Name] = c.Coax(context.TODO(), func() (interface{}, error) {
			client := ab.ClientFactory(ab.Clusters[cluster.Name])
			return maybeRetryable(client.GetRequests())
		}, "get requests from cluster %q", cluster.Name)
	}

	for cluster, promise := range promises {
		if err := promise.Err(); err != nil {
			log.Printf("Fatal: unable to get requests for cluster %q", cluster)
			return sous.NewDeployStates(), err
		}
		log.Printf("Success: got all requests from cluster %q", cluster)
		requests = append(requests, promise.Value().(dtos.SingularityRequestParentList)...)
	}

	log.Printf("Got: %d requests", len(requests))

	deployStates := sous.NewDeployStates()
	var wg sync.WaitGroup
	errChan := make(chan error)

	var counter int64

	// Start gathering all requests concurrently.
gather:
	for _, request := range requests {
		request := request
		select {
		case <-ab.Context.Done():
			log.Printf("Context ended before all deployments gathered.")
			break gather
		default:
		}

		requestID := request.Request.Id
		var requestCluster *sous.Cluster

		deployID, err := ParseRequestID(requestID)
		if err != nil {
			Log.Debug.Printf("Skipping deploy %q: unable to parse: %s", deployID, err)
			continue
		}
		oneOfMyDeploys := false
		for clusterName, cluster := range ab.Clusters {
			if deployID.Cluster == clusterName {
				requestCluster = cluster
				oneOfMyDeploys = true
				break
			}
		}
		if !oneOfMyDeploys {
			Log.Debug.Printf("Skipping deploy %q: does not match my clusters % #v", deployID, ab.Clusters)
			continue
		}

		log.Printf("Gathering data for request %q in background.", requestID)

		wg.Add(1)
		go func() {
			defer wg.Done()
			rc := ab.newRequestContext(requestID, requestCluster)
			ds, err := rc.DeployState()
			if err != nil {
				ab.ErrorCallback(err)
				errChan <- err
				return
			}
			deployStates.Add(ds)
			atomic.AddInt64(&counter, 1)
			Log.Debug.Printf("Assembled complete DeployState for request %q", requestID)

		}()
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Wait for either error or channel close.
	err := <-errChan
	Log.Debug.Printf("Gathered %d DeployStates; returning %d DeployStates.", counter, deployStates.Len())
	return deployStates, err
}

// newRequestContext initialises a requestContext and begins making HTTP
// requests to get the request (via coaxer). We can access the results of
// this via the returned requestContext's promise field.
func (ab *adsBuilder) newRequestContext(requestID string, cluster *sous.Cluster) *requestContext {
	return newRequestContext(
		ab.Context, requestID, ab.ClientFactory(cluster), *cluster, ab.Registry,
	)
}

func (ab *adsBuilder) Errorf(format string, a ...interface{}) error {
	//prefix := fmt.Sprintf("reading from cluster %q", ab.Cluster.Name)
	prefix := ""
	message := fmt.Sprintf(format, a...)
	return fmt.Errorf("%s: %s", prefix, message)
}

type temporary struct {
	error
}

func (t temporary) Temporary() bool {
	return true
}

func maybeRetryable(a interface{}, err error) (interface{}, error) {
	if err == nil {
		return a, nil
	}
	log.Printf("Maybe retryable %T? %q", err, err)
	return a, temporary{err}
}

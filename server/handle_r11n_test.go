package server

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	sous "github.com/opentable/sous/lib"
	"github.com/stretchr/testify/assert"
)

// TestNewR11nResource checks that the same queue set passed to the
// constructor makes its way to the get handler.
func TestNewR11nResource(t *testing.T) {
	qs := &sous.R11nQueueSet{}
	c := ComponentLocator{
		QueueSet: qs,
	}
	dq := newR11nResource(c)
	rm := routemap(c)

	got := dq.Get(rm, nil, &http.Request{URL: &url.URL{}}, nil).(*GETR11nHandler)
	if got.QueueSet != qs {
		t.Errorf("got different queueset")
	}
}

func TestR11nResource_Get_no_errors(t *testing.T) {

	testCases := []struct {
		desc       string
		query      string
		wantDID    sous.DeploymentID
		wantR11nID sous.R11nID
		wantWait   bool
	}{
		{
			desc:  "valid deploymentID and r11nID",
			query: "cluster=cluster1&repo=github.com%2Fuser1%2Frepo1&offset=dir1&flavor=flavor1&action=cabba9e",
			wantDID: sous.DeploymentID{
				ManifestID: sous.ManifestID{
					Source: sous.SourceLocation{
						Repo: "github.com/user1/repo1",
						Dir:  "dir1",
					},
					Flavor: "flavor1",
				},
				Cluster: "cluster1",
			},
			wantR11nID: sous.R11nID("cabba9e"),
		},
		{
			desc:  "valid short DeploymentID and r11nID",
			query: "cluster=cluster1&repo=github.com%2Fuser1%2Frepo1&action=cabba9e",
			wantDID: sous.DeploymentID{
				ManifestID: sous.ManifestID{
					Source: sous.SourceLocation{
						Repo: "github.com/user1/repo1",
					},
				},
				Cluster: "cluster1",
			},
			wantR11nID: sous.R11nID("cabba9e"),
		},
		{
			desc:  "wait query",
			query: "cluster=cluster1&repo=github.com%2Fuser1%2Frepo1&action=cabba9e&wait=true",
			wantDID: sous.DeploymentID{
				ManifestID: sous.ManifestID{
					Source: sous.SourceLocation{
						Repo: "github.com/user1/repo1",
					},
				},
				Cluster: "cluster1",
			},
			wantR11nID: sous.R11nID("cabba9e"),
			wantWait:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			dr := &R11nResource{}
			req := makeRequestWithQuery(t, tc.query)
			c := ComponentLocator{}
			rm := routemap(c)

			got := dr.Get(rm, nil, req, nil).(*GETR11nHandler)

			gotDID := got.DeploymentID
			if gotDID != tc.wantDID {
				t.Errorf("got DeploymentID:\n%#v; want:\n%#v", gotDID, tc.wantDID)
			}
			if got.DeploymentIDErr != nil {
				t.Errorf("unexpected error: %s", got.DeploymentIDErr)
			}

			gotRID := got.R11nID
			if gotRID != tc.wantR11nID {
				t.Errorf("got R11nID %q; want %q", gotRID, tc.wantR11nID)
			}

			gotWait := got.WaitForResolution
			wantWait := tc.wantWait
			if gotWait != wantWait {
				t.Errorf("got WaitForResolution %t; want %t", gotWait, wantWait)
			}
		})
	}
}

// TestGETR11nHandler_Exchange_static checks that with static queues (i.e. with
// no queue processor) we always get back nil Resolutions and sensible queue
// positions.
func TestGETR11nHandler_Exchange_static(t *testing.T) {

	queues := sous.NewR11nQueueSet()
	queuedOne1, ok := queues.Push(newR11n("one"))
	if !ok {
		t.Fatal("setup failed to push r11n")
	}
	queuedTwo1, ok := queues.Push(newR11n("two"))
	if !ok {
		t.Fatal("setup failed to push r11n")
	}
	queuedTwo2, ok := queues.Push(newR11n("two"))
	if !ok {
		t.Fatal("setup failed to push r11n")
	}

	wantNilResolution := func(t *testing.T, r r11nResponse) {
		t.Helper()
		assert.Nil(t, r.Resolution)
		if r.Resolution != (*sous.DiffResolution)(nil) {
			t.Errorf("want nil Resolution; got %#v (%[1]T) %t", r.Resolution, (r.Resolution != (*sous.DiffResolution)(nil)))
		}
	}
	wantStringReponse := func(t *testing.T, body interface{}, wantResponse string) {
		t.Helper()
		gotResponse, ok := body.(string)
		if !ok {
			t.Errorf("want a string; got a %T", body)
			return
		}
		if gotResponse != wantResponse {
			t.Errorf("got response string %q; want %q", gotResponse, wantResponse)
		}
	}
	wantStatus404 := func(t *testing.T, gotStatus int) {
		t.Helper()
		if gotStatus != 404 {
			t.Errorf("got status %d; want %d", gotStatus, 404)
		}
	}
	wantStatus200 := func(t *testing.T, gotStatus int) {
		t.Helper()
		if gotStatus != 200 {
			t.Errorf("got status %d; want %d", gotStatus, 200)
		}
	}
	wantR11nResponse := func(t *testing.T, body interface{}) r11nResponse {
		r, ok := body.(r11nResponse)
		if !ok {
			t.Fatal("got a %T; want a r11nResponse", body)
		}
		return r
	}
	wantQueuePos := func(t *testing.T, r r11nResponse, wantPos int) {
		t.Helper()
		gotPos := r.QueuePosition
		if gotPos != wantPos {
			t.Errorf("got queue position %d; want %d", gotPos, wantPos)
		}
	}
	exercise := func(depID sous.DeploymentID, r11nID sous.R11nID, wait bool) (interface{}, int) {
		gdh := &GETR11nHandler{
			QueueSet:          queues,
			DeploymentID:      depID,
			R11nID:            r11nID,
			WaitForResolution: wait,
		}
		return gdh.Exchange()
	}

	t.Run("nonexistent_deployID", func(t *testing.T) {
		t.Parallel()
		body, gotStatus := exercise(newDid("nonexistent"), "", false)
		wantStatus404(t, gotStatus)
		wantStringReponse(t, body, `Nothing queued for ":nonexistent".`)
	})

	t.Run("nonexistent_deployID_wait", func(t *testing.T) {
		t.Parallel()
		body, gotStatus := exercise(newDid("nonexistent"), "x", true)
		wantStatus404(t, gotStatus)
		wantStringReponse(t, body, `Deploy action "x" not found in queue for ":nonexistent".`)
	})

	t.Run("nonexistent_r11nID", func(t *testing.T) {
		t.Parallel()
		body, gotStatus := exercise(newDid("one"), "nonexistent", false)
		wantStatus404(t, gotStatus)
		wantStringReponse(t, body, `Deploy action "nonexistent" not found in queue for ":one".`)
	})

	t.Run("one_1", func(t *testing.T) {
		t.Parallel()
		body, gotStatus := exercise(newDid("one"), queuedOne1.ID, false)
		wantStatus200(t, gotStatus)
		gotResponse := wantR11nResponse(t, body)
		wantQueuePos(t, gotResponse, 0)
		wantNilResolution(t, gotResponse)
	})

	t.Run("two_1", func(t *testing.T) {
		t.Parallel()
		body, gotStatus := exercise(newDid("two"), queuedTwo1.ID, false)
		wantStatus200(t, gotStatus)
		gotResponse := wantR11nResponse(t, body)
		wantQueuePos(t, gotResponse, 0)
		wantNilResolution(t, gotResponse)
	})

	t.Run("two_2", func(t *testing.T) {
		t.Parallel()
		body, gotStatus := exercise(newDid("two"), queuedTwo2.ID, false)
		wantStatus200(t, gotStatus)
		gotResponse := wantR11nResponse(t, body)
		wantQueuePos(t, gotResponse, 1)
		wantNilResolution(t, gotResponse)
	})
}

// TestGETR11nHandler_Exchange_afterprocessing checks queue responses after processing.
func TestGETR11nHandler_Exchange_afterprocessing(t *testing.T) {

	queues := sous.NewR11nQueueSet(sous.R11nQueueStartWithHandler(
		func(qr *sous.QueuedR11n) sous.DiffResolution {
			return sous.DiffResolution{
				DeploymentID: qr.Rectification.Pair.ID(),
				Desc:         "test-desc",
				Error:        nil,
			}
		}))

	queuedOne1, ok := queues.Push(newR11n("one"))
	if !ok {
		t.Fatal("setup failed to push r11n")
	}
	queuedTwo1, ok := queues.Push(newR11n("two"))
	if !ok {
		t.Fatal("setup failed to push r11n")
	}
	queuedTwo2, ok := queues.Push(newR11n("two"))
	if !ok {
		t.Fatal("setup failed to push r11n")
	}

	// Wait for all rectifications to complete.
	queues.Wait(newDid("one"), queuedOne1.ID)
	queues.Wait(newDid("two"), queuedTwo1.ID)
	queues.Wait(newDid("two"), queuedTwo2.ID)

	wantStringReponse := func(t *testing.T, body interface{}, wantResponse string) {
		t.Helper()
		gotResponse, ok := body.(string)
		if !ok {
			t.Errorf("got a %T; want a string", body)
			return
		}
		if gotResponse != wantResponse {
			t.Errorf("got response string %q; want %q", gotResponse, wantResponse)
		}
	}
	wantStandardResolution := func(t *testing.T, r r11nResponse) {
		t.Helper()
		if r.Resolution == nil {
			t.Errorf("got nil resolution")
			return
		}
		if r.Resolution.Error != nil {
			t.Errorf("got resolution err %q; want nil", r.Resolution.Error)
		}
		if r.Resolution.Desc != "test-desc" {
			t.Errorf("got desc %q; want %q", r.Resolution.Desc, "test-desc")
		}
	}

	wantStatus404 := func(t *testing.T, gotStatus int) {
		t.Helper()
		if gotStatus != 404 {
			t.Errorf("got status %d; want %d", gotStatus, 404)
		}
	}
	wantStatus200 := func(t *testing.T, gotStatus int) {
		t.Helper()
		if gotStatus != 200 {
			t.Errorf("got status %d; want %d", gotStatus, 200)
		}
	}
	wantR11nResponse := func(t *testing.T, body interface{}) r11nResponse {
		r, ok := body.(r11nResponse)
		if !ok {
			t.Fatalf("got a %T (%v); want a r11nResponse", body, body)
		}
		return r
	}
	wantQueuePos := func(t *testing.T, r r11nResponse, wantPos int) {
		t.Helper()
		gotPos := r.QueuePosition
		if gotPos != wantPos {
			t.Errorf("got queue position %d; want %d", gotPos, wantPos)
		}
	}

	t.Run("nonexistent_deployID", func(t *testing.T) {
		t.Parallel()
		gdh := &GETR11nHandler{
			QueueSet:     queues,
			DeploymentID: newDid("nonexistent"),
		}
		body, gotStatus := gdh.Exchange()
		wantStatus404(t, gotStatus)
		wantStringReponse(t, body, `Nothing queued for ":nonexistent".`)
	})
	t.Run("nonexistent_deployID_wait", func(t *testing.T) {
		t.Parallel()
		gdh := &GETR11nHandler{
			QueueSet:          queues,
			DeploymentID:      newDid("nonexistent"),
			R11nID:            "x",
			WaitForResolution: true,
		}
		body, gotStatus := gdh.Exchange()
		wantStatus404(t, gotStatus)
		wantStringReponse(t, body, `Deploy action "x" not found in queue for ":nonexistent".`)
	})
	t.Run("nonexistent_r11nID", func(t *testing.T) {
		t.Parallel()
		gdh := &GETR11nHandler{
			QueueSet:     queues,
			DeploymentID: newDid("one"),
			R11nID:       "nonexistent",
		}
		body, gotStatus := gdh.Exchange()
		wantStatus404(t, gotStatus)
		wantStringReponse(t, body, `Deploy action "nonexistent" not found in queue for ":one".`)
	})
	t.Run("one_1", func(t *testing.T) {
		t.Parallel()
		gdh := &GETR11nHandler{
			QueueSet:     queues,
			DeploymentID: newDid("one"),
			R11nID:       queuedOne1.ID,
		}
		body, gotStatus := gdh.Exchange()
		wantStatus200(t, gotStatus)
		gotResponse := wantR11nResponse(t, body)
		wantQueuePos(t, gotResponse, -1)
		wantStandardResolution(t, gotResponse)
	})
	t.Run("two_1", func(t *testing.T) {
		t.Parallel()
		gdh := &GETR11nHandler{
			QueueSet:     queues,
			DeploymentID: newDid("two"),
			R11nID:       queuedTwo1.ID,
		}
		body, gotStatus := gdh.Exchange()
		wantStatus200(t, gotStatus)
		gotResponse := wantR11nResponse(t, body)
		wantQueuePos(t, gotResponse, -2)
		wantStandardResolution(t, gotResponse)
	})
	t.Run("two_2", func(t *testing.T) {
		t.Parallel()
		gdh := &GETR11nHandler{
			QueueSet:     queues,
			DeploymentID: newDid("two"),
			R11nID:       queuedTwo2.ID,
		}
		body, gotStatus := gdh.Exchange()
		wantStatus200(t, gotStatus)
		gotResponse := wantR11nResponse(t, body)
		wantQueuePos(t, gotResponse, -1)
		wantStandardResolution(t, gotResponse)
	})

	// With wait
	t.Run("one_1_wait", func(t *testing.T) {
		t.Parallel()
		gdh := &GETR11nHandler{
			QueueSet:          queues,
			DeploymentID:      newDid("one"),
			R11nID:            queuedOne1.ID,
			WaitForResolution: true,
		}
		body, gotStatus := gdh.Exchange()
		wantStatus200(t, gotStatus)
		gotResponse := wantR11nResponse(t, body)
		wantQueuePos(t, gotResponse, -1)
		wantStandardResolution(t, gotResponse)
	})
	t.Run("two_1_wait", func(t *testing.T) {
		t.Parallel()
		gdh := &GETR11nHandler{
			QueueSet:          queues,
			DeploymentID:      newDid("two"),
			R11nID:            queuedTwo1.ID,
			WaitForResolution: true,
		}
		body, gotStatus := gdh.Exchange()
		wantStatus200(t, gotStatus)
		gotResponse := wantR11nResponse(t, body)
		wantQueuePos(t, gotResponse, -2)
		wantStandardResolution(t, gotResponse)
	})
	t.Run("two_2_wait", func(t *testing.T) {
		t.Parallel()
		gdh := &GETR11nHandler{
			QueueSet:          queues,
			DeploymentID:      newDid("two"),
			R11nID:            queuedTwo2.ID,
			WaitForResolution: true,
		}
		body, gotStatus := gdh.Exchange()
		wantStatus200(t, gotStatus)
		gotResponse := wantR11nResponse(t, body)
		wantQueuePos(t, gotResponse, -1)
		wantStandardResolution(t, gotResponse)
	})
}

func TestGETR11nHandler_Exchange_errors(t *testing.T) {
	gdh := &GETR11nHandler{
		DeploymentIDErr: fmt.Errorf("this error"),
	}
	_, gotStatus := gdh.Exchange()
	const wantStatus = 404
	if gotStatus != wantStatus {
		t.Errorf("got status %d; want %d", gotStatus, wantStatus)
	}
}

// TestGETR11nHandler_Exchange_wait_success checks that when WaitForResolution
// is true, Exchange does not return until the queue reports that the r11n is
// done, and that it returns the expected Resolution result.
func TestGETR11nHandler_Exchange_wait_success(t *testing.T) {
	block := make(chan struct{})
	qh := func(qr *sous.QueuedR11n) sous.DiffResolution {
		<-block
		return sous.DiffResolution{
			DeploymentID: newDid("one"),
			Desc:         "desc1",
		}
	}
	queues := sous.NewR11nQueueSet(sous.R11nQueueStartWithHandler(qh))
	queuedOne1, ok := queues.Push(newR11n("one"))
	if !ok {
		t.Fatal("setup failed to push r11n")
	}

	grh := &GETR11nHandler{
		WaitForResolution: true,
		QueueSet:          queues,
		DeploymentID:      newDid("one"),
		R11nID:            queuedOne1.ID,
	}

	// response wraps up the pair of return parameters from Exchange.
	type response struct {
		status int
		body   interface{}
	}
	// Once Exchange completes, write its result to responses.
	responses := make(chan response)
	go func() {
		r, s := grh.Exchange()
		responses <- response{status: s, body: r}
	}()

	// At this point responses should not emit anything because block will be
	// blocking the queue from being processed, which should block Exchange from
	// returning due to grh.WaitForResolution == true.
	timeout := 10 * time.Millisecond
	select {
	// We may sometimes get away with default instead of this timeout here, but
	// the short timeout gives Exchange a little more time to complete, and
	// should result in fewer false passes for this assertion.
	case <-time.After(timeout): // OK
	case <-responses:
		t.Fatalf("Exchange returned before %s, should have waited for resolution.", timeout)
	}

	// Close block to allow the queue processor to proceed.
	close(block)
	select {
	case <-time.After(timeout):
		t.Fatalf("Exchange did not return within %s of being unblocked.", timeout)
	case got := <-responses: // OK
		const wantStatus = 200
		if got.status != wantStatus {
			t.Errorf("got status %d; want %d", got.status, wantStatus)
		}
		gotBody := got.body.(r11nResponse)
		if gotBody.Resolution == nil {
			t.Fatalf("unexpected nil Resolution")
		}
		gotDID := gotBody.Resolution.DeploymentID
		wantDID := newDid("one")
		if gotDID != wantDID {
			t.Errorf("got DeploymentID %#v; want %#v", gotDID, wantDID)
		}
		gotDesc := gotBody.Resolution.Desc
		wantDesc := sous.ResolutionType("desc1")
		if gotDesc != wantDesc {
			t.Errorf("got desc %q; want %q", gotDesc, wantDesc)
		}
	}
}

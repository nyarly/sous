package sous

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type (
	// An HTTPStateManager gets state from a Sous server and transmits updates
	// back to that server
	HTTPStateManager struct {
		serverURL *url.URL
		cached    *State
		http.Client
	}

	gdmWrapper struct {
		Deployments []*Deployment
	}
)

func (g *gdmWrapper) manifests(defs Defs) (Manifests, error) {
	ds := NewDeployments()
	for _, d := range g.Deployments {
		ds.Add(d)
	}
	return ds.Manifests(defs)
}

func (g *gdmWrapper) fromJSON(reader io.Reader) {
	dec := json.NewDecoder(reader)
	dec.Decode(g)
}

// NewHTTPStateManager creates a new HTTPStateManager
func NewHTTPStateManager(us string) (*HTTPStateManager, error) {
	u, err := url.Parse(us)
	return &HTTPStateManager{
		serverURL: u,
	}, errors.Wrapf(err, "new state manager")
}

func (hsm *HTTPStateManager) getDefs() (Defs, error) {
	ds := Defs{}
	url, err := hsm.serverURL.Parse("./defs")
	if err != nil {
		return ds, errors.Wrapf(err, "getting defs")
	}
	rq, err := hsm.Client.Get(url.String())
	if err != nil {
		return ds, errors.Wrapf(err, "getting defs")
	}

	dec := json.NewDecoder(rq.Body)

	return ds, errors.Wrapf(dec.Decode(&ds), "getting defs")
}

func (hsm *HTTPStateManager) getManifests(defs Defs) (Manifests, error) {
	url, err := hsm.serverURL.Parse("./gdm")
	if err != nil {
		return Manifests{}, errors.Wrapf(err, "getting manifests")
	}
	gdmRq, err := hsm.Client.Get(url.String())
	if err != nil {
		return Manifests{}, errors.Wrapf(err, "getting manifests")
	}
	gdm := &gdmWrapper{}
	gdm.fromJSON(gdmRq.Body)
	gdmRq.Body.Close()
	return gdm.manifests(defs)
}

// ReadState implements StateReader for HTTPStateManager
func (hsm *HTTPStateManager) ReadState() (*State, error) {
	defs, err := hsm.getDefs()
	if err != nil {
		return nil, err
	}
	ms, err := hsm.getManifests(defs)
	if err != nil {
		return nil, err
	}

	hsm.cached = &State{
		Defs:      defs,
		Manifests: ms,
	}
	return hsm.cached.Clone(), nil
}

// WriteState implements StateWriter for HTTPStateManager
func (hsm *HTTPStateManager) WriteState(ws *State) error {
	flaws := ws.Validate()
	if len(flaws) > 0 {
		return errors.Errorf("Invalid update to state: %v", flaws)
	}
	if hsm.cached == nil {
		_, err := hsm.ReadState()
		if err != nil {
			return err
		}
	}
	wds, err := ws.Deployments()
	if err != nil {
		return err
	}
	cds, err := hsm.cached.Deployments()
	if err != nil {
		return err
	}
	diff := cds.Diff(wds)
	cchs := diff.Concentrate(ws.Defs)
	return hsm.process(cchs)
}

func (hsm *HTTPStateManager) process(dc DiffConcentrator) error {
	done := make(chan struct{})
	defer close(done)

	ce := make(chan error)
	go hsm.creates(dc.Created, ce, done)

	de := make(chan error)
	go hsm.deletes(dc.Deleted, de, done)

	me := make(chan error)
	go hsm.modifies(dc.Modified, me, done)

	re := make(chan error)
	go hsm.retains(dc.Retained, re, done)

	dce := dc.Errors
	for {
		if ce == nil && de == nil && me == nil && re == nil {
			return nil
		}

		select {
		case e, open := <-dce:
			if open {
				return e
			}
			dce = nil
		case e, open := <-ce:
			if open {
				return e
			}
			ce = nil
		case e, open := <-de:
			if open {
				return e
			}
			de = nil
		case e, open := <-re:
			if open {
				return e
			}
			re = nil
		case e, open := <-me:
			if open {
				return e
			}
			me = nil
		}
	}
}

func (hsm *HTTPStateManager) manifestURL(m *Manifest) (string, error) {
	murl, err := url.Parse("./manifest")
	if err != nil {
		return "", err
	}
	mqry := url.Values{}
	mqry.Set("repo", m.Source.Repo)
	mqry.Set("offset", m.Source.Dir)
	mqry.Set("flavor", m.Flavor)
	murl.RawQuery = mqry.Encode()
	return hsm.serverURL.ResolveReference(murl).String(), nil
}

func (hsm *HTTPStateManager) manifestJSON(m *Manifest) io.Reader {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.Encode(m)
	return buf
}

func (hsm *HTTPStateManager) jsonManifest(buf io.Reader) *Manifest {
	m := &Manifest{}
	dec := json.NewDecoder(buf)
	dec.Decode(m)
	return m
}

func (hsm *HTTPStateManager) retains(mc chan *Manifest, ec chan error, done chan struct{}) {
	defer close(ec)
	for {
		select {
		case <-done:
			return
		case _, open := <-mc: //just drop 'em
			if !open {
				return
			}
		}
	}
}

func (hsm *HTTPStateManager) creates(mc chan *Manifest, ec chan error, done chan struct{}) {
	defer close(ec)
	for {
		select {
		case <-done:
			return
		case m, open := <-mc:
			if !open {
				return
			}
			if err := hsm.create(m); err != nil {
				ec <- err
			}
		}
	}
}

func (hsm *HTTPStateManager) deletes(mc chan *Manifest, ec chan error, done chan struct{}) {
	defer close(ec)
	for {
		select {
		case <-done:
			return
		case m, open := <-mc:
			if !open {
				return
			}
			if err := hsm.del(m); err != nil {
				ec <- err
			}
		}
	}
}

func (hsm *HTTPStateManager) modifies(mc chan *ManifestPair, ec chan error, done chan struct{}) {
	defer close(ec)
	for {
		select {
		case <-done:
			return
		case m, open := <-mc:
			if !open {
				return
			}
			if err := hsm.modify(m); err != nil {
				ec <- err
			}
		}
	}
}

func (hsm *HTTPStateManager) create(m *Manifest) error {
	murl, err := hsm.manifestURL(m)
	if err != nil {
		return err
	}
	rq, err := http.NewRequest("PUT", murl, hsm.manifestJSON(m))
	if err != nil {
		return errors.Wrapf(err, "create manifest request")
	}
	rq.Header.Add("If-None-Match", "*")
	rz, err := hsm.Client.Do(rq)
	if err != nil {
		return err //XXX network problems? retry?
	}
	defer rz.Body.Close()
	if rz.StatusCode != 200 {
		return errors.Errorf("%s: %#v", rz.Status, m)
	}
	return nil
}

func (hsm *HTTPStateManager) del(m *Manifest) error {
	murl, err := hsm.manifestURL(m)
	if err != nil {
		return err
	}

	grq, err := http.NewRequest("GET", murl, nil)
	if err != nil {
		return errors.Wrapf(err, "delete manifest request")
	}
	grz, err := hsm.Client.Do(grq)
	if err != nil {
		return errors.Wrapf(err, "delete manifest request")
	}
	defer grz.Body.Close()
	if !(grz.StatusCode >= 200 && grz.StatusCode < 300) {
		return errors.Errorf("GET %s to delete, %s: %#v", murl, grz.Status, m)
	}
	rm := hsm.jsonManifest(grz.Body)
	different, differences := rm.Diff(m)
	if different {
		return errors.Errorf("Remote and deleted manifests don't match: %#v", differences)
	}
	etag := grz.Header.Get("Etag")
	drq, err := http.NewRequest("DELETE", murl, nil)
	if err != nil {
		return errors.Wrapf(err, "delete manifest request")
	}
	drq.Header.Add("If-Match", etag)
	drz, err := hsm.Client.Do(drq)
	if err != nil {
		return errors.Wrapf(err, "delete manifest request")
	}
	if !(drz.StatusCode >= 200 && drz.StatusCode < 300) {
		return errors.Errorf("Delete %s failed: %s", murl, drz.Status)
	}
	return nil
}

func (hsm *HTTPStateManager) modify(mp *ManifestPair) error {
	bf := mp.Post
	af := mp.Prior
	murl, err := hsm.manifestURL(bf)
	if err != nil {
		return err
	}

	grq, err := http.NewRequest("GET", murl, nil)
	if err != nil {
		return errors.Wrapf(err, "modify request")
	}
	grz, err := hsm.Client.Do(grq)
	if err != nil {
		return errors.Wrapf(err, "modify request")
	}
	defer grz.Body.Close()
	if !(grz.StatusCode >= 200 && grz.StatusCode < 300) {
		return errors.Errorf("%s: %#v", grz.Status, bf)
	}
	rm := hsm.jsonManifest(grz.Body)
	different, differences := rm.Diff(bf)
	if different {
		return errors.Errorf("Remote and prior manifests don't match: %#v", differences)
	}
	etag := grz.Header.Get("Etag")

	murl, err = hsm.manifestURL(af)
	if err != nil {
		return err
	}

	prq, err := http.NewRequest("PUT", murl, hsm.manifestJSON(af))
	if err != nil {
		return errors.Wrapf(err, "modify request")
	}
	prq.Header.Add("If-Match", etag)
	prz, err := hsm.Client.Do(prq)
	if err != nil {
		return errors.Wrapf(err, "modify request")
	}
	if !(prz.StatusCode >= 200 && prz.StatusCode < 300) {
		return errors.Errorf("Update failed: %s / %#v", prz.Status, af)
	}
	return nil
}

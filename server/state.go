package server

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/opentable/sous/deploy"
	"github.com/opentable/sous/tools/dir"
	"github.com/opentable/sous/tools/git"
)

const (
	remote = "origin"
	branch = "master"
)

type Revision string

func (s *Server) UpdateState() (Revision, error) {
	if !dir.Exists(s.Workdir) {
		if err := git.Clone(s.Repo, s.Workdir); err != nil {
			return "", err
		}
	}
	if err := git.Forcepull(s.Workdir, remote, branch); err != nil {
		return "", err
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if err := os.Chdir(s.Workdir); err != nil {
		return "", err
	}
	info := git.GetInfo()
	if err := os.Chdir(wd); err != nil {
		return "", err
	}
	return Revision(info.CommitSHA), nil
}

func (s *Server) ReadState() (*deploy.State, error) {
	return deploy.Parse(s.Workdir)
}

func (s *Server) WriteState(to, from *deploy.State) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	*to = *from
}

func (s *Server) GetState() *deploy.State {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return &s.state
}

func (s *Server) BeginFetchingState() {
	pollFrequency := time.Duration(30) * time.Second
	log.Printf("Starting to fetch state from %s every %s", s.Repo, pollFrequency)
	for {
		revision, err := s.fetchState()
		if err != nil {
			log.Println(err, "trying again in", pollFrequency)
		} else {
			log.Printf("Successfully updated state (at %s)\n", revision)
		}
		time.Sleep(pollFrequency)
	}
}

func (s *Server) fetchState() (Revision, error) {
	revision, err := s.UpdateState()
	if err != nil {
		return "", fmt.Errorf("Unable to update state: %s", err)
	}
	state, err := s.ReadState()
	if err != nil {
		return "", fmt.Errorf("Unable to read state: %s", err)
	}
	s.WriteState(&s.state, state)
	return revision, nil
}

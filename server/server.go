package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/opentable/sous/core"
)

type Server struct {
	Repo, Workdir string
	mutex         sync.RWMutex
	state         core.State
}

func NewServer(repo string, workdir string) *Server {
	return &Server{
		Repo:    repo,
		Workdir: workdir,
		mutex:   sync.RWMutex{},
	}
}

func (s *Server) ServeMux() *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("/state", s.HandleState)
	return m
}

func (s *Server) StartStandalone(listen *url.URL) error {
	if listen.Scheme != "http" {
		return fmt.Errorf("Scheme %q not supported; want %q.", listen.Scheme, "http")
	}
	// Note url.URL.Host is either just host or host:port depending on the URL
	addr := listen.Host
	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf("%s:80", addr)
	}
	go s.BeginFetchingState()
	return http.ListenAndServe(addr, s.ServeMux())
}

func (s *Server) HandleState(w http.ResponseWriter, r *http.Request) {
	state := s.GetState()
	if state == nil {
		jsonResponse(w, map[string]string{"Status": "Waiting for state to clone..."})
	}
	jsonResponse(w, state)
}

func jsonResponse(w http.ResponseWriter, body interface{}) {
	b, err := json.MarshalIndent(body, "", "\t")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf(`{"error": %q}`, err)))
	}
	w.WriteHeader(200)
	w.Write(b)
	w.Write([]byte("\n"))
}

package commands

import (
	"net/url"
	"os"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/server"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/git"
)

func ServerHelp() string {
	return `sous server starts a standalone sous server instance
	
	note that you can embed this server easily in other Go code by using the ServeMux()
	method to get the URL handlers. That makes it easy to add your own additional endpoints, etc.`
}

func requireEnv(name string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	cli.Fatalf("You must set %s in your environment before running sous server", name)
	return ""
}

func Server(sous *core.Sous, args []string) {
	SOUS_STATE_REPO := requireEnv("SOUS_STATE_REPO")
	SOUS_STATE_WORKDIR := requireEnv("SOUS_STATE_WORKDIR")
	// Cheap validation; TODO: maybe something more obvious?
	_, err := git.ToCanonicalRepoName(SOUS_STATE_REPO)
	if err != nil {
		cli.Fatalf("Urecognised value %q for SOUS_STATE_REPO; %s", SOUS_STATE_REPO,
			err)
	}
	LISTEN_ADDR := os.Getenv("LISTEN_ADDR")
	if LISTEN_ADDR == "" {
		LISTEN_ADDR = "http://localhost:1616"
	}

	addr, err := url.Parse(LISTEN_ADDR)
	if err != nil {
		cli.Fatalf("Unable to listen on %q: %s", LISTEN_ADDR, err)
	}
	cli.Logf("Starting server at %s", LISTEN_ADDR)
	s := server.NewServer(SOUS_STATE_REPO, SOUS_STATE_WORKDIR)
	if err := s.StartStandalone(addr); err != nil {
		cli.Fatalf(err.Error())
	}
	cli.Success()
}

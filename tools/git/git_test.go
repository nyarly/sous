package git

import (
	"net/url"
	"testing"
)

func TestCanonicalName(t *testing.T) {
	urls := []string{
		"https://github.com/user/project.git",
		"http://github.com/user/project.git",
		"https://github.com/user/project",
		"http://github.com/user/project",
		"git@github.com:user/project.git",
		"git@github.com:user/project",
	}
	expected := "github.com/user/project"
	for _, s := range urls {
		u, err := url.Parse(s)
		if err != nil {
			t.Fatalf("Unable to parse url, test inconclusive: %s", u)
		}
		g := Info{OriginURL: u}
		if canonical := g.CanonicalRepoName(); canonical != expected {
			t.Errorf("Canonical name for %s was %s; want %s", s, canonical, expected)
		}
	}
}

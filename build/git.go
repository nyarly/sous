package build

import (
	"net/url"

	. "github.com/opentable/sous/util"
)

type GitInfo struct {
	CommitSHA string
	OriginURL *url.URL
}

func getGitInfo() *GitInfo {
	return &GitInfo{
		CommitSHA: Cmd("git", "rev-parse", "HEAD"),
		OriginURL: getOriginURL(),
	}
}

func (g *GitInfo) CanonicalName() string {
	return g.OriginURL.Host + g.OriginURL.Path
}

func getOriginURL() *url.URL {
	table := CmdTable("git", "remote", "-v")
	if len(table) == 0 {
		Dief("no git remotes set up")
	}
	for _, row := range table {
		if row[0] == "origin" {
			url, err := url.Parse(row[1])
			if err != nil {
				Dief("unable to parse origin (%s) as URL; %s", row[1], err)
			}
			return url
		}
	}
	Dief("unable to find remote named 'origin'")
	return nil
}

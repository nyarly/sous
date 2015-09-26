package build

import (
	"fmt"
	"net/url"

	. "github.com/opentable/sous/util"
)

type GitInfo struct {
	CommitSHA string
	OriginURL *url.URL
}

func GetGitInfo() *GitInfo {
	if err := AssertCleanWorkingTree(); err != nil {
		Dief("Unclean working tree: %s; please commit your changes", err)
	}
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

func AssertCleanWorkingTree() error {
	if IndexIsDirty() {
		return fmt.Errorf("modified files")
	}
	newFiles := UntrackedUnignoredFiles()
	if len(newFiles) == 0 {
		return nil
	}
	return fmt.Errorf("new files %v", newFiles)
}

func IndexIsDirty() bool {
	code := CmdExitCode("git", "diff-index", "--quiet", "HEAD")
	if code > 1 || code < 0 {
		Dief("Unable to determine if git index is dirty; Got exit code %d; want 0-1")
	}
	return code == 1
}

func UntrackedUnignoredFiles() []string {
	return CmdLines("git", "ls-files", "--exclude-standard", "--others")
}

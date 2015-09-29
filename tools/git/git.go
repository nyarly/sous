package git

import (
	"fmt"
	"net/url"

	. "github.com/opentable/sous/tools"
	"github.com/opentable/sous/tools/cmd"
	"github.com/opentable/sous/tools/dir"
	"github.com/opentable/sous/tools/version"
)

func RequireVersion(r *version.R) {
	if c := cmd.ExitCode("git", "--version"); c != 0 {
		Dief("git required")
	}
	v := version.Version(cmd.Table("git", "--version")[0][2])
	if !r.IsSatisfiedBy(v) {
		Dief("you have git version %s; want %s", v, r)
	}
}

func RequireRepo() {
	if !dir.Exists(".git") {
		Dief("you must be in the base of a git repository")
	}
}

type Info struct {
	CommitSHA string
	OriginURL *url.URL
}

func GetInfo() *Info {
	return &Info{
		CommitSHA: cmd.Stdout("git", "rev-parse", "HEAD"),
		OriginURL: getOriginURL(),
	}
}

func RequireCleanWorkingTree() {
	if err := AssertCleanWorkingTree(); err != nil {
		Dief("Unclean working tree: %s; please commit your changes", err)
	}
}

func (g *Info) CanonicalName() string {
	return g.OriginURL.Host + g.OriginURL.Path
}

func getOriginURL() *url.URL {
	table := cmd.Table("git", "remote", "-v")
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
	code := cmd.ExitCode("git", "diff-index", "--quiet", "HEAD")
	if code > 1 || code < 0 {
		Dief("Unable to determine if git index is dirty; Got exit code %d; want 0-1")
	}
	return code == 1
}

func UntrackedUnignoredFiles() []string {
	return cmd.New("git", "ls-files", "--exclude-standard", "--others").OutLines()
}

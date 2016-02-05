package git

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/cmd"
	"github.com/opentable/sous/tools/dir"
	"github.com/opentable/sous/tools/version"
)

type Repo struct {
	Remotes  map[string]Remote
	Branches map[string]Branch
}

type Remote struct {
	Fetch, Push *url.URL
}

type Branch struct {
	Remote, Merge string
}

func RequireVersion(r *version.R) {
	if c := cmd.ExitCode("git", "--version"); c != 0 {
		cli.Fatalf("git required")
	}
	v := version.Version(cmd.Table("git", "--version")[0][2])
	if !r.IsSatisfiedBy(v) {
		cli.Fatalf("you have git version %s; want %s", v, r.Original)
	}
}

func RequireRepo() {
	if !dir.Exists(".git") {
		cli.Fatalf("you must be in the base of a git repository")
	}
}

type Info struct {
	// The commit SHA of HEAD
	CommitSHA string
	// The URL of "origin" remote...
	// TODO: Consider following this through GitHub to find the original repo
	// if this is a fork.
	OriginURL *url.URL
	// The nearest tag before CommitSHA
	NearestTag string
	// The SHA of the nearest tag
	NearestTagSHA string
	// If there are any changed or new untracked fils, the tree is dirty
	Dirty bool
}

func GetInfo() *Info {
	nearestTagSHA := ""
	nearestTag, err := cmd.StdoutErr("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		nearestTag = ""
	} else {
		nearestTagSHA = cmd.Stdout("git", "rev-parse", nearestTag)
	}
	return &Info{
		CommitSHA:     cmd.Stdout("git", "rev-parse", "HEAD"),
		OriginURL:     getOriginURL(),
		NearestTag:    nearestTag,
		NearestTagSHA: nearestTagSHA,
	}
}

func RequireCleanWorkingTree() {
	if err := AssertCleanWorkingTree(); err != nil {
		cli.Fatalf("Unclean working tree: %s; please commit your changes", err)
	}
}

// CanonicalName returns a canonicalised version of a git repository.
// E.g. all of these:
//
//    git@github.com:user/repo.git
//    git@github.com:user/repo
//    https://github.com/user/repo.git
//    http://github.com/user/repo.git
//    https://github.com/user/repo
//    http://github.com/user/repo
//
// Become this:
//
//    github.com/user/repo
//
func (g *Info) CanonicalRepoName() string {
	// We can safely ignore the error here because ToCanonicalRepoName only
	// errors on an invalid URL, but we are using a pre-validated URL here.
	n, _ := ToCanonicalRepoName(g.OriginURL.String())
	return n
	//host := g.OriginURL.Host
	//path := g.OriginURL.Path
	//if host == "" {
	//	if !strings.ContainsRune(path, ':') {
	//		cli.Fatalf("git origin URL not recognised: %s", g.OriginURL)
	//	}
	//	p := strings.SplitN(path, ":", 2)
	//	host = p[0]
	//	path = "/" + p[1]
	//}
	//if strings.ContainsRune(host, '@') {
	//	p := strings.SplitN(host, "@", 2)
	//	host = p[1]
	//}
	//return host + strings.TrimSuffix(path, ".git")
}

func ToCanonicalRepoName(name string) (string, error) {
	u, err := url.Parse(name)
	if err != nil {
		return "", fmt.Errorf("only valid URLs can be canonicalised: %s", err)
	}
	host := u.Host
	path := u.Path
	if host == "" {
		if !strings.ContainsRune(path, ':') {
			cli.Fatalf("git origin URL not recognised: %s", u)
		}
		p := strings.SplitN(path, ":", 2)
		host = p[0]
		path = "/" + p[1]
	}
	if strings.ContainsRune(host, '@') {
		p := strings.SplitN(host, "@", 2)
		host = p[1]
	}
	return host + strings.TrimSuffix(path, ".git"), nil
}

func IsCanonicalRepoName(name string) bool {
	c, err := ToCanonicalRepoName(name)
	return err == nil && c == name
}

func getOriginURL() *url.URL {
	table := cmd.Table("git", "remote", "-v")
	if len(table) == 0 {
		cli.Fatalf("no git remotes set up")
	}
	for _, row := range table {
		if row[0] == "origin" {
			u, err := url.Parse(row[1])
			if err != nil {
				cli.Fatalf("unable to parse origin (%s) as URL; %s", row[1], err)
			}
			return u
		}
	}
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
		cli.Fatalf("Unable to determine if git index is dirty; Got exit code %d; want 0-1")
	}
	return code == 1
}

func UntrackedUnignoredFiles() []string {
	return cmd.New("git", "ls-files", "--exclude-standard", "--others").OutLines()
}

func Clone(repo, dir string) error {
	c := cmd.New("git", "clone", repo, dir)
	code := c.ExitCode()
	if code == 0 {
		return nil
	}
	defer cli.Logf("shell> %s\n%s\n%s", c, c.Stderr, c.Stdout)
	return fmt.Errorf("clone of %s into %q failed with exit code %d", repo, dir, code)
}

func Forcepull(repoDir, remote, branch string) error {
	if !dir.Exists(repoDir) {
		return fmt.Errorf("Directory %q not found", repoDir)
	}
	c := cmd.New("git", "pull", "-f", remote, branch)
	c.Setwd(repoDir)
	code := c.ExitCode()
	if code == 0 {
		return nil
	}
	defer cli.Logf("shell> %s\n%s\n%s", c, c.Stderr, c.Stdout)
	return fmt.Errorf("%s failed in %q with exit code %d", c, repoDir, code)
}

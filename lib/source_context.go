package sous

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/samsalisbury/semv"
)

type (
	// SourceContext contains contextual information about the source code being
	// built.
	SourceContext struct {
		RootDir, OffsetDir, Branch, Revision string
		Files, ModifiedFiles, NewFiles       []string
		Tags                                 []Tag
		NearestTagName, NearestTagRevision   string
		PrimaryRemoteURL                     string
		RemoteURL                            string
		RemoteURLs                           []string
		DirtyWorkingTree                     bool
		RevisionUnpushed                     bool
	}
	// Tag represents a revision control commit tag.
	Tag struct {
		Name, Revision string
	}
)

// NormalizedOffset returns a relative path from root that is based on workdir.
// Notably, it handles the case where the workdir is in the same physical path
// as root, but via symlinks
func NormalizedOffset(root, workdir string) (string, error) {
	parts := strings.Split(workdir, string(os.PathSeparator))
	for n := range parts {
		prefix := "/" + filepath.Join(parts[0:n+1]...)
		prefix, err := filepath.EvalSymlinks(prefix)
		if err != nil {
			break // this isn't working
		}
		if strings.HasPrefix(prefix, root) {
			mid := prefix[len(root):len(prefix)]
			rest := parts[n+1 : len(parts)]
			workdir = filepath.Join(append([]string{root, mid}, rest...)...)
			break
		}
	}

	relDir, err := filepath.Rel(root, workdir)
	if err != nil {
		return "", err
	}
	workdir = filepath.Join(root, relDir)
	relDir, err = filepath.Rel(root, workdir)
	if err != nil {
		return "", err
	}
	if relDir == "." {
		relDir = ""
	}
	return relDir, nil
}

// Version returns the SourceID.
func (sc *SourceContext) Version() SourceID {
	v, err := semv.Parse(sc.NearestTagName)
	if err != nil {
		v = nearestVersion(sc.Tags)
	}
	// Append revision ID.
	v = semv.MustParse(v.Format("M.m.p-?") + "+" + sc.Revision)
	sv := SourceID{
		Location: SourceLocation{
			Repo: sc.RemoteURL,
			Dir:  sc.OffsetDir,
		},
		Version: v,
	}
	Log.Debug.Printf("Version: % #v", sv)
	return sv
}

// SourceLocation returns the source location in this context.
func (sc *SourceContext) SourceLocation() SourceLocation {
	return SourceLocation{
		Repo: sc.PrimaryRemoteURL,
		Dir:  sc.OffsetDir,
	}
}

// AbsDir returns the absolute path of this source code.
func (sc *SourceContext) AbsDir() string {
	return filepath.Join(sc.RootDir, sc.OffsetDir)
}

// TagVersion returns a semver string if the most recent tag conforms to a
// semver format. Otherwise it returns an empty string
func (sc *SourceContext) TagVersion() string {
	v, err := semv.Parse(sc.NearestTagName)
	if err != nil {
		return ""
	}
	return v.Format("M.m.p")
}

func nearestVersion(tags []Tag) semv.Version {
	for _, t := range tags {
		v, err := semv.Parse(t.Name)
		if err == nil {
			return v
		}
	}
	return semv.MustParse("0.0.0-unversioned")
}

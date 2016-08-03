package sous

import (
	"path/filepath"

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
		PossiblePrimaryRemoteURL             string
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

func (sc *SourceContext) Version() SourceID {
	v, err := semv.Parse(sc.NearestTagName)
	if err != nil {
		v = nearestVersion(sc.Tags)
	}
	// Append revision ID.
	v = semv.MustParse(v.Format("M.m.p-?") + "+" + sc.Revision)
	sv := SourceID{
		RepoURL:    RepoURL(sc.RemoteURL),
		Version:    v,
		RepoOffset: RepoOffset(sc.OffsetDir),
	}
	Log.Debug.Printf("Version: % #v", sv)
	return sv
}

// AbsDir returns the absolute path of this source code.
func (sc *SourceContext) AbsDir() string {
	return filepath.Join(sc.RootDir, sc.OffsetDir)
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

package sous

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/samsalisbury/semv"
)

type (
	// BuildConfig captures the user's intent as they build a repo.
	BuildConfig struct {
		Repo, Offset, Tag, Revision string
		Strict, ForceClone          bool
		Context                     *BuildContext
	}

	// An AdvisoryName is the type for advisory tokens.
	AdvisoryName string
	// Advisories are the advisory tokens that apply to a build
	Advisories []AdvisoryName
)

const (
	// UnknownRepo is an advisory that the source workspace is not a repo.
	// TODO: Disambiguate text from NoRepoAdv, they seem like the same thing.
	UnknownRepo = AdvisoryName(`source workspace lacked repo`)
	// NoRepoAdv means there is no repository.
	// TODO: Disambiguate text from UnknownRepo.
	NoRepoAdv = AdvisoryName(`no repository`)
	// NotRequestedRevision means that a different revision was built from that
	// which was requested.
	NotRequestedRevision = AdvisoryName(`requested revision not built`)
	// Unversioned means that there was no tag at the currently checked out
	// revision, or that the tag was not a semver tag, or the tag was 0.0.0.
	Unversioned = AdvisoryName(`no versioned tag`)
	// TagMismatch means that a different tag to the one which was requested was
	// built.
	TagMismatch = AdvisoryName(`tag mismatch`)
	// TagNotHead means that the requested tag exists in the history, but there
	// were more commits since, which were part of this build.
	TagNotHead = AdvisoryName(`tag not on built revision`)
	// EphemeralTag means the tag was an ephemeral tag rather than an annotated
	// tag.
	EphemeralTag = AdvisoryName(`ephemeral tag`)
	// UnpushedRev means the revision that was build is not pushed to any
	// remote.
	UnpushedRev = AdvisoryName(`unpushed revision`)
	// BogusRev means that the revision was bogus.
	// TODO: Find out what "bogus" means.
	BogusRev = AdvisoryName(`bogus revision`)
	// DirtyWS means that the workspace was dirty, which means there were
	// untracked files present, or that one or more tracked files were modified
	// since the last commit.
	DirtyWS = AdvisoryName(`dirty workspace`)
)

// NewContext returns a new BuildContext updated based on the user's intent as expressed in the Config
func (c *BuildConfig) NewContext() *BuildContext {
	ctx := c.Context
	sc := c.Context.Source
	sh := ctx.Sh.Clone()
	sh.CD(sc.RootDir)
	bc := BuildContext{
		Sh:      sh,
		Scratch: ctx.Scratch,
		Machine: ctx.Machine,
		User:    ctx.User,
		Changes: ctx.Changes,
		Source: SourceContext{
			OffsetDir:      c.chooseOffset(),
			RemoteURL:      c.chooseRemoteURL(),
			NearestTagName: c.chooseTag(),

			RootDir:            sc.RootDir,
			Branch:             sc.Branch,
			Revision:           sc.Revision,
			Files:              sc.Files,
			ModifiedFiles:      sc.ModifiedFiles,
			NewFiles:           sc.NewFiles,
			Tags:               sc.Tags,
			NearestTagRevision: sc.NearestTagRevision,
			PrimaryRemoteURL:   sc.PrimaryRemoteURL,
			RemoteURLs:         sc.RemoteURLs,
			DirtyWorkingTree:   sc.DirtyWorkingTree,
			RevisionUnpushed:   sc.RevisionUnpushed,
		},
	}

	bc.Advisories = c.Advisories(&bc)

	return &bc
}

func (c *BuildConfig) chooseRemoteURL() string {
	if c.Repo == "" {
		Log.Debug.Printf("Using best guess: % #v", c.Context.Source.PrimaryRemoteURL)
		return c.Context.Source.PrimaryRemoteURL
	}
	return c.Repo
}

func (c *BuildConfig) chooseTag() string {
	if c.Tag == "" {
		return c.Context.Source.NearestTagName
	}
	return c.Tag
}

func (c *BuildConfig) chooseOffset() string {
	if c.Offset == "" {
		return c.Context.Source.OffsetDir
	}
	clean := filepath.Clean(c.Offset)
	if clean == "." {
		return ""
	}
	return clean
}

// Resolve settles configurations so that e.g. captured version tags are used in the absence of user input
func (c *BuildConfig) Resolve() {
	c.Tag = c.chooseTag()
}

// Validate checks that the Config is well formed
func (c *BuildConfig) Validate() error {
	if _, ve := semv.Parse(c.Tag); ve != nil {
		return fmt.Errorf("build config: tag format %q invalid: %s", c.Tag, ve)
	}
	return nil
}

// GuardStrict returns an error if there are imperfections in the proposed build
func (c *BuildConfig) GuardStrict(bc *BuildContext) error {
	if !c.Strict {
		return nil
	}
	as := bc.Advisories
	if len(as) > 0 {
		return fmt.Errorf("Strict built encountered advisories:\n  %s", strings.Join(as, "  \n"))
	}
	return nil
}

// GuardRegister returns an error if any development-only advisories exist
func (c *BuildConfig) GuardRegister(bc *BuildContext) error {
	var blockers []string
	for _, a := range bc.Advisories {
		switch AdvisoryName(a) {
		case DirtyWS, UnpushedRev, NoRepoAdv, NotRequestedRevision:
			blockers = append(blockers, a)
		}
	}
	if len(blockers) > 0 {
		return fmt.Errorf("build may not be deployable in all clusters due to advisories:\n  %s", strings.Join(blockers, "\n  "))
	}
	return nil
}

// Advisories returns a list of advisories that apply to ctx.
func (c *BuildConfig) Advisories(ctx *BuildContext) []string {
	advs := []string{}
	s := ctx.Source
	knowsRepo := false
	for _, r := range s.RemoteURLs {
		if s.RemoteURL == r {
			knowsRepo = true
			break
		}
	}
	if !knowsRepo {
		advs = append(advs, string(UnknownRepo))
	}

	if s.RemoteURL == "" {
		advs = append(advs, string(NoRepoAdv))
	}

	if c.Revision != "" && c.Revision != s.Revision {
		advs = append(advs, string(NotRequestedRevision))
	}

	if c.Context.Source.Version().Version.Format(`M.m.p`) == `0.0.0` {
		advs = append(advs, string(Unversioned))
	}

	if c.Tag != "" {
		hasTag := false
		for _, t := range s.Tags {
			if t.Name == c.Tag {
				hasTag = true
				break
			}
		}
		if !hasTag {
			advs = append(advs, string(EphemeralTag))
		} else if s.NearestTagRevision != s.Revision {
			Log.Debug.Printf("%s != %s", s.NearestTagRevision, s.Revision)
			advs = append(advs, string(TagNotHead))
		}
	}

	if s.DirtyWorkingTree {
		advs = append(advs, string(DirtyWS))
	}

	if s.RevisionUnpushed {
		advs = append(advs, string(UnpushedRev))
	}

	/*
		BuildContext struct {
			Sh      shell.Shell
			Source  SourceContext
			Scratch ScratchContext
			Machine Machine
			User    user.User
			Changes Changes
		}

		SourceContext struct {
			RootDir, OffsetDir, Branch, Revision string
			Files, ModifiedFiles, NewFiles       []string
			Tags                                 []Tag
			NearestTagName, NearestTagRevision   string
			PossiblePrimaryRemoteURL             string
			DirtyWorkingTree                     bool
		}
	*/
	return advs
}

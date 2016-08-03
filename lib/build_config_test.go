package sous

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Things that we can't easily do yet:
// ... and we're in a git workspace ...
//    If we aren't git.NewRepo() will fail
// If it's absent, we'll be building from a shallow clone
//    Again, by the time we get here, we're in a repo already
//    So, shallow clones become tricky

// If --repo is present, and we're in a git workspace, compare the --repo to
// the remotes of the workspace. If it's present, assume that we're working in
// the current workspace.

// We're now either working locally
// (in the git workspace) or in a clone.
// If --force-clone is present, we ignore
// the presence of a valid workspace and do a shallow clone anyway.

func TestPresentExplicitRepo(t *testing.T) {
	assert := assert.New(t)

	bc := BuildConfig{
		Repo: "github.com/opentable/present",
		Context: &BuildContext{
			Source: SourceContext{
				RemoteURLs: []string{
					"github.com/opentable/present",
					"github.com/opentable/also",
				},
			},
		},
	}

	ctx := bc.NewContext()
	assert.Equal(`github.com/opentable/present`, ctx.Source.RemoteURL)
}

// If it's absent, we'll be building from a shallow
// clone of the given --repo.
func TestMissingExplicitRepo(t *testing.T) {
	assert := assert.New(t)

	bc := BuildConfig{
		Repo: "github.com/opentable/present",
		Context: &BuildContext{
			Source: SourceContext{
				PossiblePrimaryRemoteURL: "github.com/guessed/upstream",
				RemoteURLs: []string{
					"github.com/opentable/also",
				},
			},
		},
	}

	ctx := bc.NewContext()
	assert.Equal(`github.com/opentable/present`, ctx.Source.RemoteURL)
	assert.Contains(ctx.Advisories, string(UnknownRepo))
}

// If --repo is absent, guess the repo from the
// remotes of the current workspace: first the upstream workspace, then the
// origin.
func TestAbsentRepoConfig(t *testing.T) {
	assert := assert.New(t)

	bc := BuildConfig{
		Repo: "",
		Context: &BuildContext{
			Source: SourceContext{
				PossiblePrimaryRemoteURL: "github.com/guessed/upstream",
				RemoteURLs: []string{
					"github.com/opentable/also",
				},
			},
		},
	}

	ctx := bc.NewContext()
	assert.Equal(`github.com/guessed/upstream`, ctx.Source.RemoteURL)
}

// If neither are present on the current workspace (or we're not in a
// git workspace), add the advisory "no repo."
func TestNoRepo(t *testing.T) {
	assert := assert.New(t)

	bc := BuildConfig{
		Repo: "",
		Context: &BuildContext{
			Source: SourceContext{
				PossiblePrimaryRemoteURL: "",
				RemoteURLs: []string{
					"github.com/opentable/also",
				},
			},
		},
	}

	ctx := bc.NewContext()
	assert.Contains(ctx.Advisories, string(NoRepoAdv))
}

// If a revision is specified, but that's not what's checked out,
// add an advisory
func TestNotRequestedRevision(t *testing.T) {
	assert := assert.New(t)

	bc := BuildConfig{
		Revision: "abcdef",
		Context: &BuildContext{
			Source: SourceContext{
				Revision: "100100100",
			},
		},
	}

	ctx := bc.NewContext()
	assert.Equal(`100100100`, ctx.Source.Revision)
	assert.Contains(ctx.Advisories, string(NotRequestedRevision))

}

func TestUsesRequestedTag(t *testing.T) {
	assert := assert.New(t)

	bc := BuildConfig{
		Tag: "1.2.3",
		Context: &BuildContext{
			Source: SourceContext{
				Revision: "abcd",
			},
		},
	}

	ctx := bc.NewContext()
	assert.Equal(`1.2.3+abcd`, ctx.Source.Version().Version.String())
}

func TestAdvisesOfDefaultVersion(t *testing.T) {
	assert := assert.New(t)

	bc := BuildConfig{
		Context: &BuildContext{
			Source: SourceContext{
				Revision: "abcd",
			},
		},
	}

	ctx := bc.NewContext()
	assert.Equal(`0.0.0-unversioned+abcd`, ctx.Source.Version().Version.String())
	assert.Contains(ctx.Advisories, string(Unversioned))
}

func TestTagNotHead(t *testing.T) {
	assert := assert.New(t)

	bc := BuildConfig{
		Tag: "1.2.3",
		Context: &BuildContext{
			Source: SourceContext{
				Revision:           "abcd",
				NearestTagName:     "1.2.3",
				NearestTagRevision: "def0",
				Tags: []Tag{
					Tag{Name: "1.2.3"},
				},
			},
		},
	}

	ctx := bc.NewContext()
	assert.Equal(`1.2.3+abcd`, ctx.Source.Version().Version.String())
	assert.Contains(ctx.Advisories, string(TagNotHead))
	assert.NotContains(ctx.Advisories, string(EphemeralTag))
}

func TestEphemeralTag(t *testing.T) {
	assert := assert.New(t)

	bc := BuildConfig{
		Tag: "1.2.3",
		Context: &BuildContext{
			Source: SourceContext{
				PossiblePrimaryRemoteURL: "github.com/opentable/present",
				Revision:                 "abcd",
				NearestTagName:           "1.2.0",
				NearestTagRevision:       "abcd",
			},
		},
	}
	/*
		bc := BuildConfig{
			Strict:   true,
			Tag:      "1.2.3",
			Repo:     "github.com/opentable/present",
			Revision: "abcdef",
			Context: &BuildContext{
				Source: SourceContext{
					RemoteURL: "github.com/opentable/present",
					RemoteURLs: []string{
						"github.com/opentable/present",
						"github.com/opentable/also",
					},
					Revision:           "abcdef",
					NearestTagName:     "1.2.3",
					NearestTagRevision: "abcdef",
					Tags: []Tag{
						Tag{Name: "1.2.3"},
					},
				},
			},
		}
	*/

	ctx := bc.NewContext()
	assert.Equal(`1.2.3+abcd`, ctx.Source.Version().Version.String())
	assert.Contains(ctx.Advisories, string(EphemeralTag))
	assert.NoError(bc.GuardRegister(ctx))
}

func TestSetsOffset(t *testing.T) {
	assert := assert.New(t)

	bc := BuildConfig{
		Offset: "sub/",
		Context: &BuildContext{
			Source: SourceContext{
				OffsetDir:          "",
				Revision:           "abcd",
				NearestTagName:     "1.2.0",
				NearestTagRevision: "def0",
			},
		},
	}

	ctx := bc.NewContext()
	assert.Equal(`sub`, ctx.Source.OffsetDir)
}

func TestDirtyWorkspaceAdvisory(t *testing.T) {
	assert := assert.New(t)

	bc := BuildConfig{
		Context: &BuildContext{
			Source: SourceContext{
				DirtyWorkingTree: true,
			},
		},
	}

	ctx := bc.NewContext()
	assert.Contains(ctx.Advisories, string(DirtyWS))
	assert.Error(bc.GuardRegister(ctx))
}

func TestUnpushedRevisionAdvisory(t *testing.T) {
	assert := assert.New(t)

	bc := BuildConfig{
		Strict: true,
		Context: &BuildContext{
			Source: SourceContext{
				RevisionUnpushed: true,
			},
		},
	}

	ctx := bc.NewContext()
	assert.Contains(ctx.Advisories, string(UnpushedRev))
	assert.Error(bc.GuardStrict(ctx))
}

func TestPermissiveGuard(t *testing.T) {
	assert := assert.New(t)

	bc := BuildConfig{
		Strict: false,
		Context: &BuildContext{
			Source: SourceContext{
				RevisionUnpushed: true,
			},
		},
	}

	ctx := bc.NewContext()
	assert.Contains(ctx.Advisories, string(UnpushedRev))
	assert.NoError(bc.GuardStrict(ctx))
}

func TestProductionReady(t *testing.T) {
	assert := assert.New(t)

	bc := BuildConfig{
		Strict:   true,
		Tag:      "1.2.3",
		Repo:     "github.com/opentable/present",
		Revision: "abcdef",
		Context: &BuildContext{
			Source: SourceContext{
				RemoteURL: "github.com/opentable/present",
				RemoteURLs: []string{
					"github.com/opentable/present",
					"github.com/opentable/also",
				},
				Revision:           "abcdef",
				NearestTagName:     "1.2.3",
				NearestTagRevision: "abcdef",
				Tags: []Tag{
					Tag{Name: "1.2.3"},
				},
			},
		},
	}

	ctx := bc.NewContext()
	assert.Len(ctx.Advisories, 0)
	assert.NoError(bc.GuardStrict(ctx))
}

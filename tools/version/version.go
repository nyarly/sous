package version

import (
	"fmt"
	"sort"
	"strings"

	"github.com/opentable/sous/tools/cli"
	"github.com/wmark/semver"
)

type V struct {
	Version  *semver.Version
	Original string
}

type R struct {
	Range    *semver.Range
	Original string
}

func Version(s string) *V {
	v, err := NewVersion(s)
	if err != nil {
		cli.Fatalf("unable to parse version string '%s'; %s", s, err)
	}
	return v
}

func NewVersion(s string) (*V, error) {
	s = strings.TrimPrefix(s, "v")
	s = strings.Replace(s, "x", "0", -1)
	s = strings.Replace(s, "X", "0", -1)
	v, err := semver.NewVersion(s)
	if err != nil {
		return nil, err
	}
	return &V{v, s}, nil
}

type VL []*V

func VersionList(vs ...string) VL {
	list := make([]*V, len(vs))
	for i, v := range vs {
		list[i] = Version(v)
	}
	return list
}

func (l VL) Strings() []string {
	s := make([]string, len(l))
	for i, v := range l {
		s[i] = v.String()
	}
	return s
}

func Range(s string) *R {
	r, err := NewRange(s)
	if err != nil {
		cli.Fatal(err)
	}
	return r
}

func NewRange(s string) (*R, error) {
	s = strings.Replace(s, "x", "0", -1)
	s = strings.Replace(s, "X", "0", -1)
	r, err := semver.NewRange(s)
	if err != nil {
		return nil, fmt.Errorf("unable to parse version range string '%s'; %s", s, err)
	}
	return &R{r, s}, nil
}

type asc []*V

func (a asc) Len() int           { return len(a) }
func (a asc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a asc) Less(i, j int) bool { return a[i].Version.Less(a[j].Version) }

func (r *R) BestMatchFrom(versions []*V) *V {
	// Sort descending so we pick the highest compatible version
	sort.Reverse(asc(versions))
	for _, v := range versions {
		if r.Range.IsSatisfiedBy(v.Version) {
			return v
		}
	}
	return nil
}

func (r *R) IsSatisfiedBy(v *V) bool {
	return r.Range.IsSatisfiedBy(v.Version)
}

func (v *V) String() string {
	return v.Original
}

func (r *R) String() string {
	return r.Original
}

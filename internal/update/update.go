// Package update checks whether a newer dewlon release exists.
package update

import (
	"cmp"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dewlonsystems/dewlon/internal/github"
)

// Repo is where dewlon releases are published.
const Repo = "dewlonsystems/dewlon"

// Status is the outcome of an update check.
type Status struct {
	Current string
	Latest  string
	URL     string
	// Comparable is false for development builds, where "newer" has no
	// clear meaning.
	Comparable bool
	Outdated   bool
}

// Check compares current against the latest published release.
func Check(ctx context.Context, current string) (Status, error) {
	rel, err := github.LatestRelease(ctx, Repo)
	if err != nil {
		return Status{}, err
	}

	st := Status{Current: current, Latest: rel.TagName, URL: rel.HTMLURL}

	cur, err := parse(current)
	if err != nil || isPseudo(cur) {
		return st, nil
	}
	c, err := Compare(current, rel.TagName)
	if err != nil {
		return st, nil
	}
	st.Comparable = true
	st.Outdated = c < 0
	return st, nil
}

type version struct {
	major, minor, patch int
	pre                 []string
}

// isPseudo spots the v0.0.0-<timestamp>-<hash> versions Go gives builds
// made from an untagged commit.
func isPseudo(v version) bool {
	return v.major == 0 && v.minor == 0 && v.patch == 0
}

func parse(s string) (version, error) {
	raw := s
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	if i := strings.IndexByte(s, '+'); i >= 0 {
		s = s[:i] // build metadata doesn't affect ordering
	}
	core, pre, _ := strings.Cut(s, "-")

	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return version{}, fmt.Errorf("%q is not a semantic version", raw)
	}

	var nums [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return version{}, fmt.Errorf("%q is not a semantic version", raw)
		}
		nums[i] = n
	}

	v := version{major: nums[0], minor: nums[1], patch: nums[2]}
	if pre != "" {
		v.pre = strings.Split(pre, ".")
	}
	return v, nil
}

// Compare returns -1, 0 or 1 as a is older than, equal to, or newer than b.
func Compare(a, b string) (int, error) {
	va, err := parse(a)
	if err != nil {
		return 0, err
	}
	vb, err := parse(b)
	if err != nil {
		return 0, err
	}

	if c := cmp.Compare(va.major, vb.major); c != 0 {
		return c, nil
	}
	if c := cmp.Compare(va.minor, vb.minor); c != 0 {
		return c, nil
	}
	if c := cmp.Compare(va.patch, vb.patch); c != 0 {
		return c, nil
	}
	return comparePre(va.pre, vb.pre), nil
}

// comparePre follows the semver rules: a release outranks any of its
// pre-releases, and identifiers compare one by one.
func comparePre(a, b []string) int {
	switch {
	case len(a) == 0 && len(b) == 0:
		return 0
	case len(a) == 0:
		return 1
	case len(b) == 0:
		return -1
	}

	for i := 0; i < len(a) && i < len(b); i++ {
		if c := compareIdent(a[i], b[i]); c != 0 {
			return c
		}
	}
	return cmp.Compare(len(a), len(b))
}

func compareIdent(a, b string) int {
	an, aErr := strconv.Atoi(a)
	bn, bErr := strconv.Atoi(b)
	switch {
	case aErr == nil && bErr == nil:
		return cmp.Compare(an, bn)
	case aErr == nil:
		return -1 // numbers sort before words
	case bErr == nil:
		return 1
	}
	return strings.Compare(a, b)
}

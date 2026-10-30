// Package source decides where a template comes from and how to prove the
// bytes we got are the ones the template's maintainers published.
//
// Releases of a template repo (tags like v1.2.0) are expected to carry two
// assets, built by the template's own release workflow:
//
//	template.zip   the template, with a single root folder
//	SHA256SUMS     checksums in the usual `sha256sum` format
//
// We download template.zip from the release and check it against SHA256SUMS.
// GitHub's automatic source archives aren't guaranteed to be byte-for-byte
// stable over time, which is why they aren't used for tags.
package source

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/dewlonsystems/dewlon/internal/downloader"
	"github.com/dewlonsystems/dewlon/internal/github"
	"github.com/dewlonsystems/dewlon/internal/templates"
)

const (
	assetName = "template.zip"
	sumsName  = "SHA256SUMS"
)

var (
	semverTag = regexp.MustCompile(`^v?\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?$`)
	sha256Hex = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// Options narrows down what to fetch.
type Options struct {
	// Ref is a branch, tag or commit. Empty means the template's default.
	Ref string
	// SHA256 pins the expected checksum of the download.
	SHA256 string
	// SkipVerify downloads a tag's plain source archive without checking
	// it against published checksums.
	SkipVerify bool
}

// Plan is a resolved download.
type Plan struct {
	Request downloader.Request
	Ref     string
	// Want is the checksum the download must match. Empty means the
	// download can't be verified (a branch, for example).
	Want string
}

// Checked reports whether the download will be verified.
func (p Plan) Checked() bool { return p.Want != "" }

// Resolve works out the URL to fetch and the checksum to expect.
func Resolve(ctx context.Context, tpl templates.Template, opts Options) (Plan, error) {
	ref := opts.Ref
	if ref == "" {
		ref = tpl.DefaultRef
	}

	var pin string
	if opts.SHA256 != "" {
		var err error
		if pin, err = normalizeSHA(opts.SHA256); err != nil {
			return Plan{}, err
		}
	}

	token := github.Token()

	var plan Plan
	if !opts.SkipVerify && semverTag.MatchString(ref) {
		var err error
		if plan, err = releasePlan(ctx, tpl, ref, token); err != nil {
			return Plan{}, err
		}
	} else {
		plan = Plan{Request: archiveRequest(tpl, ref, token), Ref: ref}
	}

	if pin != "" {
		if plan.Want != "" && plan.Want != pin {
			return Plan{}, fmt.Errorf("--sha256 %s does not match the checksum published for %s (%s)", pin, ref, plan.Want)
		}
		plan.Want = pin
	}
	return plan, nil
}

// Verify compares the checksum of what was downloaded with what we expect.
func Verify(p Plan, got string) error {
	if p.Want == "" {
		return nil
	}
	if !strings.EqualFold(p.Want, got) {
		return fmt.Errorf("checksum mismatch for %s: expected %s but downloaded %s; the download was discarded",
			p.Ref, p.Want, got)
	}
	return nil
}

func releasePlan(ctx context.Context, tpl templates.Template, ref, token string) (Plan, error) {
	rel, err := github.ReleaseByTag(ctx, tpl.Repo, ref)
	if err != nil {
		return Plan{}, fmt.Errorf("looking up release %s of %s: %w", ref, tpl.Repo, err)
	}

	archive, okArchive := rel.Asset(assetName)
	sums, okSums := rel.Asset(sumsName)
	if !okArchive || !okSums {
		return Plan{}, fmt.Errorf(
			"release %s of %s has no %s and %s, so it can't be verified; pass --no-verify to use the plain source archive instead",
			ref, tpl.Repo, assetName, sumsName)
	}

	sumsURL, sumsAccept := assetLocation(sums, token)
	data, err := github.GetBytes(ctx, sumsURL, sumsAccept, 1<<20)
	if err != nil {
		return Plan{}, fmt.Errorf("downloading %s: %w", sumsName, err)
	}
	want, err := ParseSums(data, assetName)
	if err != nil {
		return Plan{}, fmt.Errorf("reading %s for %s: %w", sumsName, ref, err)
	}

	archiveURL, archiveAccept := assetLocation(archive, token)
	return Plan{
		Request: downloader.Request{
			URL:    archiveURL,
			Accept: archiveAccept,
			Token:  github.TokenFor(archiveURL),
		},
		Ref:  ref,
		Want: want,
	}, nil
}

// assetLocation picks the URL to download a release asset from. With a
// token we go through the API, which also works for private repos; without
// one the public download link avoids spending API quota.
func assetLocation(a github.Asset, token string) (rawURL, accept string) {
	if token != "" {
		return a.URL, "application/octet-stream"
	}
	return a.BrowserDownloadURL, "application/octet-stream"
}

func archiveRequest(tpl templates.Template, ref, token string) downloader.Request {
	if token != "" {
		return downloader.Request{
			URL:   fmt.Sprintf("https://api.github.com/repos/%s/zipball/%s", tpl.Repo, ref),
			Token: token,
		}
	}
	return downloader.Request{URL: tpl.ArchiveURL(ref)}
}

// ParseSums finds name in a `sha256sum`-style listing ("<hex>  <file>" or
// "<hex> *<file>") and returns its checksum in lowercase hex.
func ParseSums(data []byte, name string) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}
		file := strings.TrimPrefix(strings.TrimPrefix(fields[1], "*"), "./")
		if file != name {
			continue
		}
		sum := strings.ToLower(fields[0])
		if !sha256Hex.MatchString(sum) {
			return "", fmt.Errorf("the checksum listed for %s isn't a valid SHA-256", name)
		}
		return sum, nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("no checksum listed for %s", name)
}

func normalizeSHA(s string) (string, error) {
	s = strings.ToLower(strings.TrimPrefix(strings.ToLower(strings.TrimSpace(s)), "sha256:"))
	if !sha256Hex.MatchString(s) {
		return "", fmt.Errorf("invalid --sha256 value: expected 64 hex characters")
	}
	return s, nil
}

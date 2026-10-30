// Package github is a small client for the few GitHub endpoints dewlon uses.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dewlonsystems/dewlon/internal/buildinfo"
)

const apiBase = "https://api.github.com"

var client = &http.Client{Timeout: 30 * time.Second}

// tokenVars are checked in order. DEWLON_GITHUB_TOKEN wins so people can
// give dewlon a narrowly scoped token without touching their usual one.
var tokenVars = []string{"DEWLON_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"}

// Token returns the GitHub token from the environment, or "" if none is set.
func Token() string {
	for _, name := range tokenVars {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			return v
		}
	}
	return ""
}

// TokenFor returns the token only when rawURL points at GitHub, so it can
// never leak into a request to some other server.
func TokenFor(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	switch u.Hostname() {
	case "api.github.com", "github.com":
		return Token()
	}
	return ""
}

// Asset is a file attached to a release.
type Asset struct {
	Name               string `json:"name"`
	URL                string `json:"url"` // API URL; works for private repos with a token
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// Release is the part of a GitHub release dewlon cares about.
type Release struct {
	TagName    string  `json:"tag_name"`
	HTMLURL    string  `json:"html_url"`
	Prerelease bool    `json:"prerelease"`
	Draft      bool    `json:"draft"`
	Assets     []Asset `json:"assets"`
}

// Asset finds an attached file by name.
func (r Release) Asset(name string) (Asset, bool) {
	for _, a := range r.Assets {
		if a.Name == name {
			return a, true
		}
	}
	return Asset{}, false
}

// APIError is returned when GitHub answers with something other than 200.
type APIError struct {
	Status   int
	URL      string
	HadToken bool
}

func (e *APIError) Error() string {
	switch e.Status {
	case http.StatusNotFound:
		if e.HadToken {
			return fmt.Sprintf("GitHub returned 404 for %s; check the name, and that your token can read the repository", e.URL)
		}
		return fmt.Sprintf("GitHub returned 404 for %s; if the repository is private, set DEWLON_GITHUB_TOKEN", e.URL)
	case http.StatusUnauthorized:
		return "GitHub rejected the token (401); check DEWLON_GITHUB_TOKEN"
	case http.StatusForbidden, http.StatusTooManyRequests:
		return fmt.Sprintf("GitHub refused the request (%d), most likely a rate limit; setting DEWLON_GITHUB_TOKEN raises it", e.Status)
	}
	return fmt.Sprintf("GitHub returned %d for %s", e.Status, e.URL)
}

// LatestRelease fetches the newest published release of repo ("owner/name").
func LatestRelease(ctx context.Context, repo string) (Release, error) {
	return fetchRelease(ctx, fmt.Sprintf("%s/repos/%s/releases/latest", apiBase, repo))
}

// ReleaseByTag fetches the release published for tag.
func ReleaseByTag(ctx context.Context, repo, tag string) (Release, error) {
	return fetchRelease(ctx, fmt.Sprintf("%s/repos/%s/releases/tags/%s", apiBase, repo, url.PathEscape(tag)))
}

func fetchRelease(ctx context.Context, rawURL string) (Release, error) {
	data, err := GetBytes(ctx, rawURL, "", 5<<20)
	if err != nil {
		return Release{}, err
	}
	var rel Release
	if err := json.Unmarshal(data, &rel); err != nil {
		return Release{}, fmt.Errorf("reading GitHub response: %w", err)
	}
	return rel, nil
}

// GetBytes downloads a small resource into memory. Anything larger than
// limit is rejected. An empty accept header asks for GitHub's JSON API.
func GetBytes(ctx context.Context, rawURL, accept string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	if accept == "" {
		accept = "application/vnd.github+json"
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("User-Agent", "dewlon/"+buildinfo.Get().Version)

	token := TokenFor(rawURL)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{Status: resp.StatusCode, URL: rawURL, HadToken: token != ""}
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("response from %s is larger than %d bytes", rawURL, limit)
	}
	return data, nil
}

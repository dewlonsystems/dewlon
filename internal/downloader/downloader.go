// Package downloader fetches files over HTTP and fingerprints them.
package downloader

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dewlonsystems/dewlon/internal/buildinfo"
)

const (
	maxAttempts = 3

	// MaxSize caps how much we'll download. Templates are small; anything
	// near this size is almost certainly the wrong URL.
	MaxSize int64 = 200 << 20
)

var client = &http.Client{Timeout: 3 * time.Minute}

// Request describes one download.
type Request struct {
	URL string
	// Accept sets the Accept header. GitHub's API needs
	// application/octet-stream to hand back a release asset.
	Accept string
	// Token, if set, is sent as a bearer token.
	Token string
}

// statusError carries the HTTP status so we can decide whether a retry
// makes sense.
type statusError struct {
	code     int
	status   string
	url      string
	hadToken bool
}

func (e *statusError) Error() string {
	switch e.code {
	case http.StatusNotFound:
		msg := fmt.Sprintf("not found (404) at %s; check the template name and --ref", e.url)
		if !e.hadToken && isGitHub(e.url) {
			msg += ". If the repository is private, set DEWLON_GITHUB_TOKEN"
		}
		return msg
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Sprintf("access denied (%d) for %s; check that DEWLON_GITHUB_TOKEN can read the repository", e.code, e.url)
	}
	return fmt.Sprintf("server replied %s for %s", e.status, e.url)
}

func isGitHub(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && strings.HasSuffix(u.Hostname(), "github.com")
}

// Download saves the response to dest and returns its SHA-256 as hex.
// Network hiccups and 5xx replies are retried a couple of times with a
// short backoff; 4xx replies fail at once since trying again won't help.
func Download(ctx context.Context, r Request, dest string) (string, error) {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var sum string
		sum, lastErr = fetch(ctx, r, dest)
		if lastErr == nil {
			return sum, nil
		}
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if !retryable(lastErr) || attempt == maxAttempts {
			break
		}

		wait := time.Second << (attempt - 1) // 1s, then 2s
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(wait):
		}
	}

	return "", lastErr
}

func fetch(ctx context.Context, r Request, dest string) (sum string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.URL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "dewlon/"+buildinfo.Get().Version)
	if r.Accept != "" {
		req.Header.Set("Accept", r.Accept)
	}
	if r.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.Token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	// Nothing useful can be done about a failed close on a response body
	// we've already read.
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", &statusError{code: resp.StatusCode, status: resp.Status, url: r.URL, hadToken: r.Token != ""}
	}
	if resp.ContentLength > MaxSize {
		return "", fmt.Errorf("download is %d bytes, over the %d byte limit", resp.ContentLength, MaxSize)
	}

	file, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	// Don't leave a partial file behind, and don't ignore a failed close
	// here: on some filesystems that's where a write error finally shows up.
	defer func() {
		if cerr := file.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			_ = os.Remove(dest)
		}
	}()

	hash := sha256.New()
	n, err := io.Copy(io.MultiWriter(file, hash), io.LimitReader(resp.Body, MaxSize+1))
	if err != nil {
		return "", err
	}
	if n > MaxSize {
		return "", fmt.Errorf("download exceeded the %d byte limit", MaxSize)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func retryable(err error) bool {
	var se *statusError
	if errors.As(err, &se) {
		return se.code >= 500 || se.code == http.StatusTooManyRequests || se.code == http.StatusRequestTimeout
	}
	// Anything else is a transport problem: dropped connection, DNS, timeout.
	return true
}

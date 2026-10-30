// Package downloader fetches template archives over HTTP.
package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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

// statusError carries the HTTP status so we can decide whether a retry
// makes sense.
type statusError struct {
	code   int
	status string
	url    string
}

func (e *statusError) Error() string {
	if e.code == http.StatusNotFound {
		return fmt.Sprintf("not found (404) at %s; check the template name and --ref", e.url)
	}
	return fmt.Sprintf("server replied %s for %s", e.status, e.url)
}

// Download saves url to the file at dest. Network hiccups and 5xx replies
// are retried a couple of times with a short backoff; 4xx replies fail at
// once since trying again won't change the answer.
func Download(ctx context.Context, url, dest string) error {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		lastErr = fetch(ctx, url, dest)
		if lastErr == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !retryable(lastErr) || attempt == maxAttempts {
			break
		}

		wait := time.Second << (attempt - 1) // 1s, then 2s
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}

	return lastErr
}

func fetch(ctx context.Context, url, dest string) (err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "dewlon/"+buildinfo.Version)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &statusError{code: resp.StatusCode, status: resp.Status, url: url}
	}
	if resp.ContentLength > MaxSize {
		return fmt.Errorf("download is %d bytes, over the %d byte limit", resp.ContentLength, MaxSize)
	}

	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	// Don't leave a partial file behind, and don't ignore a failed close:
	// on some filesystems that's where a write error finally shows up.
	defer func() {
		if cerr := file.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			_ = os.Remove(dest)
		}
	}()

	n, err := io.Copy(file, io.LimitReader(resp.Body, MaxSize+1))
	if err != nil {
		return err
	}
	if n > MaxSize {
		return fmt.Errorf("download exceeded the %d byte limit", MaxSize)
	}
	return nil
}

func retryable(err error) bool {
	var se *statusError
	if errors.As(err, &se) {
		return se.code >= 500 || se.code == http.StatusTooManyRequests || se.code == http.StatusRequestTimeout
	}
	// Anything else is a transport problem: dropped connection, DNS, timeout.
	return true
}

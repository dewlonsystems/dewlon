// Package runner runs external commands such as `go mod tidy`.
package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// MissingToolError means the program isn't installed or isn't on PATH.
type MissingToolError struct {
	Name string
}

func (e *MissingToolError) Error() string {
	return fmt.Sprintf("%s was not found on your PATH", e.Name)
}

// Run executes name with args inside dir. Output is always captured so a
// failure can show what went wrong; if stream is not nil, the output is
// also copied there as it arrives (used for --verbose).
func Run(ctx context.Context, dir string, stream io.Writer, name string, args ...string) error {
	path, err := exec.LookPath(name)
	if err != nil {
		return &MissingToolError{Name: name}
	}

	var captured bytes.Buffer
	var out io.Writer = &captured
	if stream != nil {
		out = io.MultiWriter(&captured, stream)
	}

	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Dir = dir
	cmd.Stdout = out
	cmd.Stderr = out

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("%s %s failed: %w\n%s",
			name, strings.Join(args, " "), err, tail(captured.String(), 2000))
	}
	return nil
}

// tail keeps the last n bytes; the end of a tool's output is where the
// useful part usually is.
func tail(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return "..." + s[len(s)-n:]
}

// Package unzip extracts zip archives safely.
package unzip

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Limits guard against archives that try to fill the disk.
type Limits struct {
	MaxFiles      int
	MaxTotalBytes int64 // total uncompressed size
}

// DefaultLimits is generous for any real template.
var DefaultLimits = Limits{
	MaxFiles:      20_000,
	MaxTotalBytes: 500 << 20,
}

// Extract unpacks zipFile into dest. It refuses entries that would land
// outside dest ("zip slip"), rejects symlinks, and enforces the limits on
// what is actually written rather than trusting the sizes in the headers.
func Extract(ctx context.Context, zipFile, dest string, limits Limits) error {
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("opening archive: %w", err)
	}
	// The archive is only read, so a failed close loses nothing.
	defer func() { _ = reader.Close() }()

	if len(reader.File) > limits.MaxFiles {
		return fmt.Errorf("archive has %d entries, over the limit of %d", len(reader.File), limits.MaxFiles)
	}

	root, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	var written int64
	for _, f := range reader.File {
		if err := ctx.Err(); err != nil {
			return err
		}

		target, err := safeJoin(root, f.Name)
		if err != nil {
			return err
		}

		if f.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("archive entry %q is a symlink, which isn't allowed", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}

		n, err := extractFile(f, target, limits.MaxTotalBytes-written)
		if err != nil {
			return err
		}
		written += n
	}

	return nil
}

// safeJoin joins name onto root and makes sure the result is still inside it.
func safeJoin(root, name string) (string, error) {
	target := filepath.Join(root, filepath.FromSlash(name))

	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("archive entry %q escapes the destination folder", name)
	}
	return target, nil
}

func extractFile(f *zip.File, target string, budget int64) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return 0, err
	}

	src, err := f.Open()
	if err != nil {
		return 0, err
	}
	defer func() { _ = src.Close() }()

	// Keep the original permissions (so shell scripts stay executable),
	// but never let a file end up unreadable to its owner.
	perm := f.Mode().Perm()
	if perm == 0 {
		perm = 0o644
	}
	perm |= 0o600

	dst, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return 0, err
	}

	n, copyErr := io.Copy(dst, io.LimitReader(src, budget+1))
	closeErr := dst.Close()

	switch {
	case copyErr != nil:
		return n, copyErr
	case closeErr != nil:
		return n, closeErr
	case n > budget:
		return n, fmt.Errorf("archive is larger than the %d byte limit once unpacked", budget)
	}
	return n, nil
}

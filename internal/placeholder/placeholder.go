// Package placeholder fills in {{TOKENS}} inside a freshly extracted template.
package placeholder

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Files bigger than this are skipped; no source file is this large.
const maxFileSize = 5 << 20

// leftover matches tokens that look like ours, so we can point out any the
// template used but we had no value for.
var leftover = regexp.MustCompile(`\{\{[A-Z][A-Z0-9_]*\}\}`)

// Apply replaces {{KEY}} with its value for every key in values, in file
// contents and in file and folder names. Binary files and .git are left
// alone, and files are only rewritten when something actually changed.
//
// It returns any {{UPPER_CASE}} tokens that were still present afterwards.
func Apply(root string, values map[string]string) (unresolved []string, err error) {
	replacer := newReplacer(values)
	seen := map[string]struct{}{}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if d.Name() == ".git" && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Size() > maxFileSize {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if isBinary(data) {
			return nil
		}

		original := string(data)
		updated := replacer.Replace(original)

		for _, tok := range leftover.FindAllString(updated, -1) {
			seen[tok] = struct{}{}
		}

		if updated == original {
			return nil
		}
		return os.WriteFile(path, []byte(updated), info.Mode().Perm())
	})
	if err != nil {
		return nil, fmt.Errorf("filling placeholders in files: %w", err)
	}

	if err := renamePaths(root, replacer); err != nil {
		return nil, fmt.Errorf("filling placeholders in names: %w", err)
	}

	for tok := range seen {
		unresolved = append(unresolved, tok)
	}
	sort.Strings(unresolved)
	return unresolved, nil
}

// newReplacer builds a single-pass replacer. Keys are sorted so the result
// never depends on map iteration order.
func newReplacer(values map[string]string) *strings.Replacer {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		pairs = append(pairs, "{{"+k+"}}", values[k])
	}
	return strings.NewReplacer(pairs...)
}

// renamePaths renames files and folders whose names contain a token.
func renamePaths(root string, replacer *strings.Replacer) error {
	var paths []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() && d.Name() == ".git" && path != root {
			return filepath.SkipDir
		}
		if path != root && strings.Contains(d.Name(), "{{") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// A child path is always longer than its parent's, so going longest
	// first renames children before the folders that contain them.
	sort.Slice(paths, func(i, j int) bool { return len(paths[i]) > len(paths[j]) })

	for _, path := range paths {
		dir, base := filepath.Split(path)
		renamed := replacer.Replace(base)
		if renamed == base {
			continue
		}
		if strings.ContainsAny(renamed, `/\`) {
			return fmt.Errorf("cannot rename %q to %q: a name can't contain a path separator", base, renamed)
		}

		target := filepath.Join(dir, renamed)
		if _, err := os.Lstat(target); err == nil {
			return fmt.Errorf("cannot rename %q: %q already exists", path, target)
		}
		if err := os.Rename(path, target); err != nil {
			return err
		}
	}
	return nil
}

// isBinary uses the same trick git does: a NUL byte near the start means
// it isn't text.
func isBinary(data []byte) bool {
	n := len(data)
	if n > 8000 {
		n = 8000
	}
	return bytes.IndexByte(data[:n], 0) >= 0
}

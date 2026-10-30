// Package filesystem holds small file helpers used while scaffolding.
package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
)

// holdingName is where the archive's root folder waits while we move its
// contents up a level.
const holdingName = ".dewlon-root"

// Flatten takes a folder that holds exactly one subfolder (the way GitHub
// zips unpack, e.g. fiber-template-main/) and moves that subfolder's
// contents up into dir.
func Flatten(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	if len(entries) != 1 || !entries[0].IsDir() {
		return fmt.Errorf("expected a single root folder after extraction, found %d entries", len(entries))
	}

	// Park the root under a neutral name first. Otherwise a template that
	// contains a folder named like its own root would collide on the way up.
	holding := filepath.Join(dir, holdingName)
	if err := os.Rename(filepath.Join(dir, entries[0].Name()), holding); err != nil {
		return err
	}

	children, err := os.ReadDir(holding)
	if err != nil {
		return err
	}

	for _, child := range children {
		from := filepath.Join(holding, child.Name())
		to := filepath.Join(dir, child.Name())
		if err := os.Rename(from, to); err != nil {
			return err
		}
	}

	return os.Remove(holding)
}

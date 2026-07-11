package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
)

func Flatten(projectDir string) error {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return err
	}

	if len(entries) != 1 || !entries[0].IsDir() {
		return fmt.Errorf("expected a single root directory after extraction")
	}

	root := filepath.Join(projectDir, entries[0].Name())

	files, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	for _, file := range files {
		oldPath := filepath.Join(root, file.Name())
		newPath := filepath.Join(projectDir, file.Name())

		if err := os.Rename(oldPath, newPath); err != nil {
			return err
		}
	}

	return os.Remove(root)
}

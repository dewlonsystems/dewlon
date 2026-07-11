package template

import (
	"os"
	"path/filepath"
	"strings"
)

func Replace(root string, replacements map[string]string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		text := string(content)

		for old, newValue := range replacements {
			text = strings.ReplaceAll(text, old, newValue)
		}

		return os.WriteFile(path, []byte(text), info.Mode())
	})
}

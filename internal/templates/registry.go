// Package templates is the catalogue of project templates dewlon can scaffold.
package templates

import (
	"fmt"
	"strings"
)

// Template describes one scaffold and where it lives.
type Template struct {
	// Key is what people type: dewlon new app -t <key>.
	Key         string
	Name        string
	Description string

	// Repo is the GitHub repository as "owner/name".
	Repo string
	// DefaultRef is the branch, tag or commit used when --ref isn't given.
	DefaultRef string

	// Install is the command run inside the new project to fetch its
	// dependencies. Leave it empty if the template needs none.
	Install []string
	// NextSteps are shown to the user once the project is ready.
	NextSteps []string
}

// ArchiveURL returns the zip download link for the given ref, falling back
// to the template's default when ref is empty.
func (t Template) ArchiveURL(ref string) string {
	if ref == "" {
		ref = t.DefaultRef
	}
	return fmt.Sprintf("https://github.com/%s/archive/%s.zip", t.Repo, ref)
}

// registry is a slice rather than a map so the order stays stable in
// listings and in the interactive picker. The first entry is the default.
var registry = []Template{
	{
		Key:         "fiber",
		Name:        "Fiber API",
		Description: "Go REST API built on Fiber",
		Repo:        "dewlonsystems/fiber-template",
		DefaultRef:  "main",
		Install:     []string{"go", "mod", "tidy"},
		NextSteps:   []string{"go run ."},
	},
	{
		Key:         "flutter",
		Name:        "Flutter app",
		Description: "Flutter application",
		Repo:        "dewlonsystems/flutter-template",
		DefaultRef:  "main",
		Install:     []string{"flutter", "pub", "get"},
		NextSteps:   []string{"flutter run"},
	},
}

// All returns every template, default first. The slice is a copy, so
// callers can't change the registry by accident.
func All() []Template {
	out := make([]Template, len(registry))
	copy(out, registry)
	return out
}

// Keys returns the template keys in registry order.
func Keys() []string {
	keys := make([]string, 0, len(registry))
	for _, t := range registry {
		keys = append(keys, t.Key)
	}
	return keys
}

// Get looks a template up by key, ignoring case and surrounding spaces.
func Get(key string) (Template, error) {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, t := range registry {
		if t.Key == key {
			return t, nil
		}
	}
	return Template{}, fmt.Errorf("unknown template %q (available: %s)", key, strings.Join(Keys(), ", "))
}

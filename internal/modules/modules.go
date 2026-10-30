// Package modules holds add-ons that can be dropped into an existing Go
// project with `dewlon add`. Their files are embedded in the binary, so
// adding one works offline and always matches the CLI's version.
package modules

import (
	"embed"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed all:files
var embedded embed.FS

// Module describes one add-on.
type Module struct {
	Key         string
	Description string
	// Requires are Go modules to fetch with `go get` after the files land.
	Requires []string
	// Notes are shown once the module is added.
	Notes []string
}

var registry = []Module{
	{
		Key:         "docker",
		Description: "Multi-stage Dockerfile and .dockerignore for a small production image",
		Notes:       []string{"Build the image with: docker build -t <name> ."},
	},
	{
		Key:         "ci",
		Description: "GitHub Actions workflow that vets and tests every push and pull request",
	},
	{
		Key:         "database",
		Description: "PostgreSQL connection pool (pgx) configured from DATABASE_URL",
		Requires:    []string{"github.com/jackc/pgx/v5"},
		Notes: []string{
			"Set DATABASE_URL, e.g. postgres://user:pass@localhost:5432/app",
			"Call database.Connect(ctx) from main and close the pool on shutdown.",
		},
	},
}

// All returns every module.
func All() []Module {
	out := make([]Module, len(registry))
	copy(out, registry)
	return out
}

// Keys returns the module keys in registry order.
func Keys() []string {
	keys := make([]string, 0, len(registry))
	for _, m := range registry {
		keys = append(keys, m.Key)
	}
	return keys
}

// Get looks a module up by key.
func Get(key string) (Module, error) {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, m := range registry {
		if m.Key == key {
			return m, nil
		}
	}
	return Module{}, fmt.Errorf("unknown module %q (available: %s)", key, strings.Join(Keys(), ", "))
}

// Project is what we learn about the project a module is added to.
type Project struct {
	Dir         string // absolute path
	Name        string
	Module      string // Go module path from go.mod
	GoVersion   string
	MainPackage string // e.g. "." or "./cmd/api"
}

// Inspect reads go.mod in dir and works out the details modules need.
func Inspect(dir string) (Project, error) {
	if dir == "" {
		dir = "."
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Project{}, err
	}

	data, err := os.ReadFile(filepath.Join(abs, "go.mod"))
	if errors.Is(err, fs.ErrNotExist) {
		return Project{}, fmt.Errorf("no go.mod in %s; run this from the root of a Go project or pass --dir", abs)
	}
	if err != nil {
		return Project{}, err
	}

	p := Project{Dir: abs, Name: filepath.Base(abs), GoVersion: "1"}
	for _, line := range strings.Split(string(data), "\n") {
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "module "):
			p.Module = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "module ")), `"`)
		case strings.HasPrefix(line, "go "):
			p.GoVersion = strings.TrimSpace(strings.TrimPrefix(line, "go "))
		}
	}
	if p.Module == "" {
		return Project{}, fmt.Errorf("%s/go.mod has no module line", abs)
	}

	p.MainPackage = detectMain(abs)
	return p, nil
}

// detectMain finds the package to build: the root if it's a main package,
// otherwise the only folder under cmd/ that is one. When it can't tell, it
// falls back to the root.
func detectMain(dir string) string {
	if hasMainPackage(dir) {
		return "."
	}

	entries, err := os.ReadDir(filepath.Join(dir, "cmd"))
	if err != nil {
		return "."
	}

	var found []string
	for _, e := range entries {
		if e.IsDir() && hasMainPackage(filepath.Join(dir, "cmd", e.Name())) {
			found = append(found, "./cmd/"+e.Name())
		}
	}
	if len(found) == 1 {
		return found[0]
	}
	return "."
}

func hasMainPackage(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.PackageClauseOnly)
		if err == nil && f.Name.Name == "main" {
			return true
		}
	}
	return false
}

// File is one file a module would write.
type File struct {
	Path   string // slash-separated, relative to the project
	Data   []byte
	Mode   fs.FileMode
	Exists bool // already present in the project
}

// Plan renders the module's files for p without writing anything.
func Plan(m Module, p Project) ([]File, error) {
	root := "files/" + m.Key
	values := strings.NewReplacer(
		"{{PROJECT_NAME}}", p.Name,
		"{{MODULE_NAME}}", p.Module,
		"{{GO_VERSION}}", p.GoVersion,
		"{{MAIN_PACKAGE}}", p.MainPackage,
		"{{YEAR}}", fmt.Sprint(time.Now().Year()),
	)

	var out []File
	err := fs.WalkDir(embedded, root, func(name string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}

		data, err := embedded.ReadFile(name)
		if err != nil {
			return err
		}

		// Files are stored with a .tmpl suffix so Go tooling and editors
		// don't treat the embedded sources as part of this module.
		rel := strings.TrimSuffix(strings.TrimPrefix(name, root+"/"), ".tmpl")
		_, statErr := os.Stat(filepath.Join(p.Dir, filepath.FromSlash(rel)))

		out = append(out, File{
			Path:   rel,
			Data:   []byte(values.Replace(string(data))),
			Mode:   0o644,
			Exists: statErr == nil,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("reading module %q: %w", m.Key, err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("module %q has no files", m.Key)
	}
	return out, nil
}

// ConflictError lists files that already exist and would be overwritten.
type ConflictError struct {
	Paths []string
}

func (e *ConflictError) Error() string {
	return "these files already exist: " + strings.Join(e.Paths, ", ")
}

// Apply writes the files into the project. Existing files are only
// replaced when force is set. If anything fails part-way, files written so
// far are put back the way they were.
func Apply(p Project, files []File, force bool) error {
	var conflicts []string
	for _, f := range files {
		if f.Exists {
			conflicts = append(conflicts, f.Path)
		}
	}
	if len(conflicts) > 0 && !force {
		return &ConflictError{Paths: conflicts}
	}

	type undo struct {
		path     string
		original []byte
		existed  bool
	}
	var done []undo

	rollback := func() {
		for i := len(done) - 1; i >= 0; i-- {
			u := done[i]
			if u.existed {
				_ = os.WriteFile(u.path, u.original, 0o644)
			} else {
				_ = os.Remove(u.path)
			}
		}
	}

	for _, f := range files {
		target := filepath.Join(p.Dir, filepath.FromSlash(f.Path))

		u := undo{path: target}
		if orig, err := os.ReadFile(target); err == nil {
			u.existed, u.original = true, orig
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			rollback()
			return fmt.Errorf("creating folder for %s: %w", f.Path, err)
		}
		if err := os.WriteFile(target, f.Data, f.Mode); err != nil {
			rollback()
			return fmt.Errorf("writing %s: %w", f.Path, err)
		}
		done = append(done, u)
	}
	return nil
}

package modules

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newProject(t *testing.T) Project {
	t.Helper()
	dir := t.TempDir()

	write := func(rel, body string) {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module github.com/acme/billing // comment\n\ngo 1.25\n")
	write("cmd/api/main.go", "package main\n\nfunc main() {}\n")

	p, err := Inspect(dir)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestInspect(t *testing.T) {
	p := newProject(t)
	if p.Module != "github.com/acme/billing" || p.GoVersion != "1.25" || p.MainPackage != "./cmd/api" {
		t.Errorf("unexpected project: %+v", p)
	}
}

func TestInspectWithoutGoMod(t *testing.T) {
	if _, err := Inspect(t.TempDir()); err == nil {
		t.Error("expected an error when go.mod is missing")
	}
}

func TestPlanDockerFillsPlaceholders(t *testing.T) {
	p := newProject(t)
	mod, _ := Get("docker")

	files, err := Plan(mod, p)
	if err != nil {
		t.Fatal(err)
	}

	var dockerfile string
	for _, f := range files {
		if f.Path == "Dockerfile" {
			dockerfile = string(f.Data)
		}
	}
	for _, want := range []string{"golang:1.25-alpine", "./cmd/api", "/out/" + p.Name} {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("Dockerfile is missing %q:\n%s", want, dockerfile)
		}
	}
	if strings.Contains(dockerfile, "{{") {
		t.Errorf("unfilled placeholder left in Dockerfile:\n%s", dockerfile)
	}
}

func TestEveryModuleHasFiles(t *testing.T) {
	p := newProject(t)
	for _, m := range All() {
		if _, err := Plan(m, p); err != nil {
			t.Errorf("module %q: %v", m.Key, err)
		}
	}
}

func TestApplyRefusesToOverwrite(t *testing.T) {
	p := newProject(t)
	mod, _ := Get("docker")

	files, err := Plan(mod, p)
	if err != nil {
		t.Fatal(err)
	}
	if err := Apply(p, files, false); err != nil {
		t.Fatal(err)
	}

	// A second run now collides with what the first one wrote.
	files, _ = Plan(mod, p)
	err = Apply(p, files, false)

	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected a ConflictError, got %v", err)
	}
	if err := Apply(p, files, true); err != nil {
		t.Errorf("--force should overwrite: %v", err)
	}
}

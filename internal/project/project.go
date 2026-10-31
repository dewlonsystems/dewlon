// Package project builds a new project on disk from a template.
package project

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dewlonsystems/dewlon/internal/downloader"
	"github.com/dewlonsystems/dewlon/internal/filesystem"
	"github.com/dewlonsystems/dewlon/internal/placeholder"
	"github.com/dewlonsystems/dewlon/internal/runner"
	"github.com/dewlonsystems/dewlon/internal/source"
	"github.com/dewlonsystems/dewlon/internal/templates"
	"github.com/dewlonsystems/dewlon/internal/unzip"
)

// defaultModulePrefix is used to build the Go module path when none is given.
const defaultModulePrefix = "github.com/dewlonsystems/"

// Options tunes Create. The zero value is fine: current folder, template's
// default ref, install dependencies, say nothing.
type Options struct {
	// Dir is the folder the project is created in. Defaults to ".".
	Dir string
	// Module overrides the Go module path (fills {{MODULE_NAME}}).
	Module string
	// Ref is the branch, tag or commit to build from. Defaults to the
	// template's own default.
	Ref string
	// SHA256 pins the checksum the downloaded template must have.
	SHA256 string
	// SkipVerify turns off checksum verification for tagged releases.
	SkipVerify bool
	// SkipInstall leaves dependencies alone, for offline use or CI.
	SkipInstall bool

	// Logf, if set, receives a line for each step.
	Logf func(format string, args ...any)
	// InstallOutput, if set, gets the dependency installer's live output.
	InstallOutput io.Writer
}

// Result describes what was created.
type Result struct {
	// Path is the absolute path of the new project.
	Path     string
	Template templates.Template
	Ref      string
	// Verified is true when the download was checked against a checksum.
	Verified bool
	// Warnings are problems that didn't stop the project from being
	// created, such as dependencies that couldn't be installed.
	Warnings []string
}

// Create scaffolds projectName from the template identified by templateKey.
//
// The project is assembled in a hidden staging folder next to its final
// location and only renamed into place once it is complete, so a failure
// or Ctrl+C never leaves a half-built project under the real name.
//
// Installing dependencies is the one step that can fail without failing
// Create: the scaffold is already good by then, and a flaky network is no
// reason to throw it away. That failure comes back as a warning.
func Create(ctx context.Context, projectName, templateKey string, opts Options) (res *Result, err error) {
	logf := opts.Logf
	if logf == nil {
		logf = func(string, ...any) {}
	}

	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if err = ValidateName(projectName); err != nil {
		return nil, err
	}

	tpl, err := templates.Get(templateKey)
	if err != nil {
		return nil, err
	}

	module := opts.Module
	if module == "" {
		module = defaultModulePrefix + projectName
	} else if err = validateModule(module); err != nil {
		return nil, err
	}

	ref := opts.Ref
	if ref == "" {
		ref = tpl.DefaultRef
	} else if err = validateRef(ref); err != nil {
		return nil, err
	}

	parent := opts.Dir
	if parent == "" {
		parent = "."
	}
	info, statErr := os.Stat(parent)
	switch {
	case errors.Is(statErr, fs.ErrNotExist):
		return nil, fmt.Errorf("destination folder %q does not exist", parent)
	case statErr != nil:
		return nil, fmt.Errorf("checking destination folder %q: %w", parent, statErr)
	case !info.IsDir():
		return nil, fmt.Errorf("destination %q is not a folder", parent)
	}

	target := filepath.Join(parent, projectName)
	if _, statErr = os.Lstat(target); statErr == nil {
		return nil, fmt.Errorf("%q already exists; choose a different name or remove it first", target)
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return nil, fmt.Errorf("checking %q: %w", target, statErr)
	}

	// Work out where the template comes from (and what checksum to expect)
	// before touching the disk, so a bad --ref or --sha256 fails fast.
	plan, err := source.Resolve(ctx, tpl, source.Options{
		Ref:        ref,
		SHA256:     opts.SHA256,
		SkipVerify: opts.SkipVerify,
	})
	if err != nil {
		return nil, err
	}

	// The download lives in the system temp folder so it never clutters the
	// user's directory.
	zipFile, err := os.CreateTemp("", "dewlon-*.zip")
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}
	zipPath := zipFile.Name()
	_ = zipFile.Close()
	defer func() { _ = os.Remove(zipPath) }()

	staging, err := os.MkdirTemp(parent, ".dewlon-"+projectName+"-")
	if err != nil {
		return nil, fmt.Errorf("creating staging folder: %w", err)
	}
	// After a successful rename this path no longer exists, so removing it
	// is a harmless no-op on the happy path.
	defer func() { _ = os.RemoveAll(staging) }()

	// If anything fails or is cancelled after the project is in place,
	// take it away again so the user is left with a clean slate.
	created := false
	defer func() {
		if err != nil && created {
			_ = os.RemoveAll(target)
		}
	}()

	logf("Downloading %s (%s)", tpl.Name, ref)
	sum, err := downloader.Download(ctx, plan.Request, zipPath)
	if err != nil {
		return nil, fmt.Errorf("downloading template: %w", err)
	}
	if err = source.Verify(plan, sum); err != nil {
		return nil, err
	}
	if plan.Checked() {
		logf("Checksum verified (sha256:%s)", sum[:12])
	} else {
		logf("Checksum not verified (%s is a moving reference; use a release tag or --sha256 to pin it)", ref)
	}

	logf("Unpacking")
	if err = unzip.Extract(ctx, zipPath, staging, unzip.DefaultLimits); err != nil {
		return nil, fmt.Errorf("unpacking template: %w", err)
	}
	if err = filesystem.Flatten(staging); err != nil {
		return nil, fmt.Errorf("organizing project: %w", err)
	}

	logf("Customizing")
	unresolved, err := placeholder.Apply(staging, map[string]string{
		"PROJECT_NAME": projectName,
		"APP_NAME":     projectName,
		"MODULE_NAME":  module,
		"PACKAGE_NAME": packageName(projectName),
		"YEAR":         fmt.Sprint(time.Now().Year()),
	})
	if err != nil {
		return nil, fmt.Errorf("customizing template: %w", err)
	}

	var warnings []string
	if len(unresolved) > 0 {
		warnings = append(warnings,
			"the template left these placeholders unfilled: "+strings.Join(unresolved, ", "))
	}

	// MkdirTemp makes folders private (0700); a project folder shouldn't be.
	_ = os.Chmod(staging, 0o755)
	if err = os.Rename(staging, target); err != nil {
		return nil, fmt.Errorf("moving project into place: %w", err)
	}
	created = true

	// Dependencies are installed after the move on purpose: tools like
	// `flutter pub get` write absolute paths into the project.
	if !opts.SkipInstall && len(tpl.Install) > 0 {
		cmdline := strings.Join(tpl.Install, " ")
		logf("Installing dependencies (%s)", cmdline)

		runErr := runner.Run(ctx, target, opts.InstallOutput, tpl.Install[0], tpl.Install[1:]...)
		if ctx.Err() != nil {
			err = ctx.Err()
			return nil, err
		}
		if runErr != nil {
			var missing *runner.MissingToolError
			if errors.As(runErr, &missing) {
				warnings = append(warnings, fmt.Sprintf(
					"%s isn't installed, so dependencies were skipped; once it is, run '%s' inside the project",
					missing.Name, cmdline))
			} else {
				warnings = append(warnings, fmt.Sprintf(
					"dependencies were not installed; run '%s' inside the project\n%v", cmdline, runErr))
			}
		}
	}

	abs, absErr := filepath.Abs(target)
	if absErr != nil {
		abs = target
	}

	return &Result{
		Path:     abs,
		Template: tpl,
		Ref:      ref,
		Verified: plan.Checked(),
		Warnings: warnings,
	}, nil
}

// packageName turns "my-app" into "my_app", which is what Dart (and other
// ecosystems that dislike hyphens) expect. A leading digit gets a prefix
// because package names can't start with one.
func packageName(name string) string {
	s := strings.ToLower(strings.ReplaceAll(name, "-", "_"))
	if s[0] >= '0' && s[0] <= '9' {
		s = "app_" + s
	}
	return s
}

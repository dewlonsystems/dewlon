package project

import (
	"fmt"
	"os"

	"github.com/dewlonsystems/dewlon/internal/downloader"
	"github.com/dewlonsystems/dewlon/internal/filesystem"
	"github.com/dewlonsystems/dewlon/internal/runner"
	"github.com/dewlonsystems/dewlon/internal/template"
	"github.com/dewlonsystems/dewlon/internal/templates"
	"github.com/dewlonsystems/dewlon/internal/unzip"
)

// Options controls Create's behavior beyond the project name and template.
type Options struct {
	// Verbose, if true, prints each step as it happens (download, extract,
	// organize, customize, install). Default is silent: nothing prints
	// until Create returns, success or failure.
	Verbose bool
}

// Create scaffolds a new project from the named template.
//
// On any failure, Create removes anything it already wrote — the
// downloaded zip and, if project creation got far enough to create the
// target directory, that directory too — so a failed run never leaves a
// half-built project directory behind for the caller to clean up by hand.
func Create(projectName, templateName string, opts Options) (err error) {
	logf := func(format string, args ...any) {
		if opts.Verbose {
			fmt.Printf(format+"\n", args...)
		}
	}

	// zipPath is unique per invocation (based on the project name) rather
	// than a fixed "template.zip" in the current directory, so two
	// concurrent or interrupted runs can't collide on the same file.
	zipPath := ".dewlon-" + projectName + ".zip"

	// dirCreated tracks whether the target project directory now exists,
	// so the cleanup path below only removes it if this run is actually
	// responsible for creating it.
	dirCreated := false

	defer func() {
		_ = os.Remove(zipPath)
		if err != nil && dirCreated {
			_ = os.RemoveAll(projectName)
		}
	}()

	// Refuse to overwrite an existing directory/file — this is a cheap,
	// important guard: without it, a typo'd project name could silently
	// extract a template on top of existing, unrelated files.
	if _, statErr := os.Stat(projectName); statErr == nil {
		return fmt.Errorf("cannot create project: %q already exists", projectName)
	} else if !os.IsNotExist(statErr) {
		return fmt.Errorf("checking project path %q: %w", projectName, statErr)
	}

	tpl, err := templates.Get(templateName)
	if err != nil {
		return fmt.Errorf("resolving template %q: %w", templateName, err)
	}

	if err = downloader.Download(tpl.URL, zipPath); err != nil {
		return fmt.Errorf("downloading template: %w", err)
	}
	logf("Downloaded %s template", tpl.Name)

	if err = unzip.Extract(zipPath, projectName); err != nil {
		return fmt.Errorf("extracting template: %w", err)
	}
	dirCreated = true
	logf("Extracted template")

	if err = filesystem.Flatten(projectName); err != nil {
		return fmt.Errorf("organizing project: %w", err)
	}
	logf("Organized project")

	replacements := map[string]string{
		"{{PROJECT_NAME}}": projectName,
		"{{MODULE_NAME}}":  "github.com/dewlonsystems/" + projectName,
		"{{APP_NAME}}":     projectName,
	}
	if err = template.Replace(projectName, replacements); err != nil {
		return fmt.Errorf("customizing template: %w", err)
	}
	logf("Customized template")

	if err = runner.GoModTidy(projectName); err != nil {
		return fmt.Errorf("installing dependencies: %w", err)
	}
	logf("Installed dependencies")

	return nil
}

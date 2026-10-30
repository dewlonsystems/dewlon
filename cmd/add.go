/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dewlonsystems/dewlon/internal/modules"
	"github.com/dewlonsystems/dewlon/internal/runner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	addDir       string
	addForce     bool
	addDryRun    bool
	addNoInstall bool
)

var addCmd = &cobra.Command{
	Use:   "add [module]",
	Short: "Add a module to an existing Go project",
	Long: `Adds a ready-made module (Dockerfile, CI workflow, database setup, ...)
to the Go project in the current folder. Run it from the folder that holds
go.mod, or point at it with --dir.

Run it without a module name to see what's available. Nothing is
overwritten unless you pass --force, and --dry-run shows what would happen.

Examples:

  dewlon add
  dewlon add docker
  dewlon add database --dry-run
  dewlon add ci --dir ../billing-api`,
	Args: cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var out []string
		for _, m := range modules.All() {
			out = append(out, m.Key+"\t"+m.Description)
		}
		return out, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return listModules()
	}

	mod, err := modules.Get(args[0])
	if err != nil {
		return err
	}

	proj, err := modules.Inspect(addDir)
	if err != nil {
		return err
	}

	files, err := modules.Plan(mod, proj)
	if err != nil {
		return err
	}

	if addDryRun {
		say("Adding %q to %s would write:", mod.Key, proj.Name)
		for _, f := range files {
			label := color.GreenString("create   ")
			if f.Exists {
				label = color.YellowString("overwrite")
			}
			say("  %s %s", label, f.Path)
		}
		return nil
	}

	if err := modules.Apply(proj, files, addForce); err != nil {
		var conflict *modules.ConflictError
		if errors.As(err, &conflict) {
			return fmt.Errorf("%w; re-run with --force to overwrite them, or --dry-run to preview", err)
		}
		return err
	}

	success("Added %s to %s", mod.Key, proj.Name)
	for _, f := range files {
		say("  %s %s", color.GreenString("+"), f.Path)
	}

	if !addNoInstall && len(mod.Requires) > 0 {
		installRequires(cmd, proj, mod)
	}

	if len(mod.Notes) > 0 {
		say("")
		for _, note := range mod.Notes {
			say("  %s", note)
		}
	}
	return nil
}

// installRequires fetches the module's Go dependencies. The files are
// already in place by now, so a failure here is a warning, not an error.
func installRequires(cmd *cobra.Command, proj modules.Project, mod modules.Module) {
	var stream *os.File
	if verbose && !quiet {
		stream = os.Stdout
	}

	run := func(args ...string) error {
		if stream == nil {
			return runner.Run(cmd.Context(), proj.Dir, nil, "go", args...)
		}
		return runner.Run(cmd.Context(), proj.Dir, stream, "go", args...)
	}

	say("")
	say("Fetching dependencies...")

	for _, req := range mod.Requires {
		if err := run("get", req+"@latest"); err != nil {
			warnInstall(err, "go get "+req)
			return
		}
	}
	if err := run("mod", "tidy"); err != nil {
		warnInstall(err, "go mod tidy")
	}
}

func warnInstall(err error, manual string) {
	var missing *runner.MissingToolError
	if errors.As(err, &missing) {
		warn("go isn't installed, so dependencies were skipped; once it is, run '%s'", manual)
		return
	}
	warn("dependencies were not fetched; run '%s' yourself\n%v", manual, err)
}

func listModules() error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "MODULE\tDESCRIPTION")
	for _, m := range modules.All() {
		fmt.Fprintf(w, "%s\t%s\n", m.Key, m.Description)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	say("")
	say("Add one with: dewlon add <module>")
	return nil
}

func init() {
	flags := addCmd.Flags()
	flags.StringVarP(&addDir, "dir", "d", ".", "project folder (the one containing go.mod)")
	flags.BoolVar(&addForce, "force", false, "overwrite files that already exist")
	flags.BoolVar(&addDryRun, "dry-run", false, "show what would be written without changing anything")
	flags.BoolVar(&addNoInstall, "no-install", false, "don't fetch Go dependencies afterwards")

	_ = addCmd.RegisterFlagCompletionFunc("dir",
		func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveFilterDirs
		})

	rootCmd.AddCommand(addCmd)
}

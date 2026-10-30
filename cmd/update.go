/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dewlonsystems/dewlon/internal/buildinfo"
	"github.com/dewlonsystems/dewlon/internal/update"
	"github.com/spf13/cobra"
)

var failIfOutdated bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check whether a newer dewlon release is available",
	Long: `Looks up the latest dewlon release on GitHub and tells you whether
yours is behind. It never changes anything on your machine; it just points
you at the right upgrade command for the way dewlon was installed.

Use --fail-if-outdated in scripts or CI to turn "out of date" into a
non-zero exit status.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 20*time.Second)
		defer cancel()

		current := buildinfo.Get().Version

		st, err := update.Check(ctx, current)
		if err != nil {
			return fmt.Errorf("checking for updates: %w", err)
		}

		switch {
		case !st.Comparable:
			say("This is a development build (%s). The latest release is %s.", st.Current, st.Latest)

		case st.Outdated:
			say("A newer version is available: %s (you have %s).", st.Latest, st.Current)
			say("Release notes: %s", st.URL)
			say("")
			say("To upgrade, run:")
			say("  %s", upgradeHint(executablePath(), runtime.GOOS))
			if failIfOutdated {
				return errors.New("dewlon is out of date")
			}

		default:
			success("dewlon %s is up to date", st.Current)
		}
		return nil
	},
}

// executablePath returns where the running binary lives, with symlinks
// resolved, or "" if that can't be worked out.
func executablePath() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		return resolved
	}
	return exe
}

// upgradeHint guesses how dewlon was installed from where the binary sits
// and returns the matching upgrade command.
func upgradeHint(exe, goos string) string {
	p := strings.ToLower(filepath.ToSlash(exe))

	switch {
	case strings.Contains(p, "/cellar/"), strings.Contains(p, "/caskroom/"), strings.Contains(p, "/homebrew/"):
		return "brew upgrade dewlon"
	case strings.Contains(p, "/scoop/"):
		return "scoop update dewlon"
	case strings.Contains(p, "/go/bin/"):
		return "go install github.com/dewlonsystems/dewlon@latest"
	case goos == "windows":
		return "irm https://raw.githubusercontent.com/dewlonsystems/dewlon/main/scripts/install.ps1 | iex"
	}
	return "curl -fsSL https://raw.githubusercontent.com/dewlonsystems/dewlon/main/scripts/install.sh | sh"
}

func init() {
	updateCmd.Flags().BoolVar(&failIfOutdated, "fail-if-outdated", false, "exit with an error when a newer release exists")
	rootCmd.AddCommand(updateCmd)
}

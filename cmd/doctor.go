/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var strict bool

// tool is something dewlon or the projects it creates can make use of.
type tool struct {
	name    string
	args    []string // how to ask the tool for its version
	why     string
	install string
}

var doctorTools = []tool{
	{
		name:    "go",
		args:    []string{"version"},
		why:     "installs dependencies for the fiber template",
		install: "https://go.dev/dl",
	},
	{
		name:    "flutter",
		args:    []string{"--version"},
		why:     "installs dependencies for the flutter template",
		install: "https://docs.flutter.dev/get-started/install",
	},
	{
		name:    "git",
		args:    []string{"--version"},
		why:     "handy for versioning the project you create",
		install: "https://git-scm.com/downloads",
	},
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check that your machine is ready to create and run dewlon projects",
	Long: `Checks for the tools dewlon projects use and whether templates can be
downloaded. Missing tools are reported as warnings, since you may only need
some of them. Use --strict to exit with an error if anything is missing.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		problems := 0

		for _, t := range doctorTools {
			version, err := probe(ctx, t)
			if err == nil {
				fmt.Printf("  %s %-8s %s\n", color.GreenString("✓"), t.name, version)
				continue
			}
			problems++
			fmt.Printf("  %s %-8s not found (%s)\n", color.YellowString("!"), t.name, t.why)
			fmt.Printf("    install: %s\n", t.install)
		}

		if err := checkNetwork(ctx); err != nil {
			problems++
			fmt.Printf("  %s %-8s can't reach github.com (%v)\n", color.YellowString("!"), "network", err)
			fmt.Println("    templates are downloaded from GitHub, so creating projects will fail")
		} else {
			fmt.Printf("  %s %-8s github.com is reachable\n", color.GreenString("✓"), "network")
		}

		fmt.Println()
		if problems == 0 {
			success("everything looks good")
			return nil
		}
		if strict {
			return errors.New("some checks did not pass")
		}
		return nil
	},
}

// probe finds the tool on PATH and returns the first line of its version
// output. The timeout stops a broken install from hanging the command.
func probe(ctx context.Context, t tool) (string, error) {
	path, err := exec.LookPath(t.name)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, path, t.args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("running %s: %w", t.name, err)
	}

	line, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	return strings.TrimSpace(line), nil
}

func checkNetwork(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, "https://github.com", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func init() {
	doctorCmd.Flags().BoolVar(&strict, "strict", false, "exit with an error if any check fails (useful in CI)")
	rootCmd.AddCommand(doctorCmd)
}

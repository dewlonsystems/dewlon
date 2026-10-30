/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dewlonsystems/dewlon/internal/buildinfo"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Global flags, available on every subcommand.
var (
	quiet   bool
	verbose bool
	noColor bool
)

var rootCmd = &cobra.Command{
	Use:   "dewlon",
	Short: "Scaffold new projects from Dewlon's maintained templates",
	Long: `Dewlon is a command line tool for starting new projects from Dewlon's
maintained templates (Fiber APIs, Flutter apps, and more to come).

Examples:

  dewlon new hospital-api
  dewlon new school-api --template=fiber
  dewlon new mobile-app -t flutter
  dewlon templates
  dewlon doctor`,

	Version: buildinfo.Get().Version,

	// We print errors ourselves in Execute, and a full usage dump after
	// every failure is more noise than help.
	SilenceUsage:  true,
	SilenceErrors: true,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			color.NoColor = true
		}
	},
}

// Execute runs the CLI. It is called once, from main.
//
// Ctrl+C (or SIGTERM) cancels the context handed to every command, so
// long-running work can stop and clean up after itself.
func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	err := rootCmd.ExecuteContext(ctx)
	stop()

	if err == nil {
		return
	}

	if errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, "\ncancelled")
		os.Exit(130) // the usual exit code for an interrupted process
	}

	fail(err)
	hint("run 'dewlon --help' for usage")
	os.Exit(1)
}

func init() {
	rootCmd.SetVersionTemplate("dewlon {{.Version}}\n")

	flags := rootCmd.PersistentFlags()
	flags.BoolVarP(&verbose, "verbose", "v", false, "show each step as it happens")
	flags.BoolVarP(&quiet, "quiet", "q", false, "only print errors")
	flags.BoolVar(&noColor, "no-color", false, "turn off colored output")
}

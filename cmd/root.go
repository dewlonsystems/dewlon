/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "dewlon",
	Short: "Scaffold new Go projects from Dewlon templates",
	Long: `Dewlon is a CLI for scaffolding new projects from Dewlon's
maintained templates (Fiber APIs, Flutter apps, and more).

Examples:

  dewlon new hospital-api
  dewlon new school-api --template=fiber
  dewlon new
  dewlon version`,
}

// Execute adds all child commands to the root command and sets flags
// appropriately. This is called by main.main(). It only needs to happen
// once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// No persistent or local flags on the root command itself right now.
	// Subcommands (new, version, ...) define their own flags.
}

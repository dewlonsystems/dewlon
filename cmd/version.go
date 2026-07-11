/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the dewlon CLI version. Update this per release, or inject a
// value at build time via -ldflags "-X github.com/dewlonsystems/dewlon/cmd.Version=v0.2.0"
// so a tagged release always reports the tag it was built from.
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the dewlon version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("dewlon version", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

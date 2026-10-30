/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import (
	"fmt"

	"github.com/dewlonsystems/dewlon/internal/buildinfo"
	"github.com/spf13/cobra"
)

var shortVersion bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and build details",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		info := buildinfo.Get()

		if shortVersion {
			fmt.Println(info.Version)
			return
		}

		fmt.Printf("dewlon %s\n", info.Version)
		fmt.Printf("  commit:   %s\n", info.Commit)
		fmt.Printf("  built:    %s\n", info.Date)
		fmt.Printf("  go:       %s\n", info.GoVersion)
		fmt.Printf("  platform: %s\n", info.Platform)
	},
}

func init() {
	versionCmd.Flags().BoolVar(&shortVersion, "short", false, "print only the version number")
	rootCmd.AddCommand(versionCmd)
}

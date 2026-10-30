/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dewlonsystems/dewlon/internal/templates"
	"github.com/spf13/cobra"
)

var templatesCmd = &cobra.Command{
	Use:     "templates",
	Aliases: []string{"list", "ls"},
	Short:   "List the available project templates",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "TEMPLATE\tDESCRIPTION\tSOURCE")
		for _, t := range templates.All() {
			fmt.Fprintf(w, "%s\t%s\t%s@%s\n", t.Key, t.Description, t.Repo, t.DefaultRef)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(templatesCmd)
}

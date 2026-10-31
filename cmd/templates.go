/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import (
	"github.com/dewlonsystems/dewlon/internal/templates"
	"github.com/spf13/cobra"
)

var templatesCmd = &cobra.Command{
	Use:     "templates",
	Aliases: []string{"list", "ls"},
	Short:   "List the available project templates",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var rows [][]string
		for _, t := range templates.All() {
			rows = append(rows, []string{t.Key, t.Description, t.Repo + "@" + t.DefaultRef})
		}
		return printTable([]string{"TEMPLATE", "DESCRIPTION", "SOURCE"}, rows)
	},
}

func init() {
	rootCmd.AddCommand(templatesCmd)
}

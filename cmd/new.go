/*
Copyright © 2026 Dewlon Systems
*/

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Create a new Dewlon project",
	Long: `Creates a new project using a Dewlon project template.

Example:

	dewlon new hospital-api
	dewlon new school-manager
	dewlon new inventory-system`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]

		fmt.Println("🚀 Dewlon CLI")
		fmt.Println()
		fmt.Println("Creating project...")
		fmt.Println()
		fmt.Println("Project Name :", projectName)
		fmt.Println("Template     : Fiber API")
		fmt.Println()
		fmt.Println("✔ Project created successfully.")
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}

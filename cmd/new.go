/*
Copyright © 2026 Dewlon Systems
*/

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/dewlonsystems/dewlon/internal/project"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	templateName string
	verbose      bool
)

// validName restricts project names to safe, portable characters —
// letters, numbers, hyphens, and underscores, starting with a letter or
// digit. This avoids path traversal (e.g. "../../etc"), spaces, and other
// characters that would produce an invalid Go module or directory name.
var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// allowedTemplates is the fixed set of templates newCmd accepts. Keeping
// this as a validated list means an invalid --template value fails fast,
// right here, with a clear message — instead of failing deep inside
// project.Create with whatever error that layer happens to produce.
var allowedTemplates = []string{"fiber", "flutter"}

// printLogo prints the Dewlon "D" mark: a two-tone ASCII wordmark, most of
// it in the terminal's default color with a red accent stripe. Purely
// cosmetic — never gates behavior, and degrades harmlessly on terminals
// without color support (fatih/color falls back to plain text there).
func printLogo() {
	red := color.New(color.FgRed).SprintFunc()

	fmt.Println(" ██████╗  " + red("███████╗"))
	fmt.Println(" ██╔══██╗ " + red("██╔════╝"))
	fmt.Println(" ██║  ██║ " + red("█████╗  "))
	fmt.Println(" ██║  ██║ " + red("██╔══╝  "))
	fmt.Println(" ██████╔╝ " + red("███████╗"))
	fmt.Println(" ╚═════╝  " + red("╚══════╝"))
	fmt.Println()
}

// promptProjectName interactively asks the user for a project name when
// none was given on the command line, re-prompting on empty input or names
// that fail validName. It reads from stdin directly rather than relying on
// any argument parsing, so it works the same regardless of how newCmd was
// invoked.
func promptProjectName() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Project name: ")

		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("reading project name: %w", err)
		}

		name := strings.TrimSpace(line)

		if name == "" {
			fmt.Println("Project name cannot be empty.")
			continue
		}
		if !validName.MatchString(name) {
			fmt.Println("Invalid name: use letters, numbers, hyphens, and underscores only, starting with a letter or number.")
			continue
		}

		return name, nil
	}
}

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Create a new Dewlon project",
	Long: `Creates a new project using one of the available Dewlon templates.

If project-name is omitted, you'll be prompted for one.

Examples:

  dewlon new hospital-api
  dewlon new school-api --template=fiber
  dewlon new mobile-app --template=flutter
  dewlon new`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		printLogo()

		var projectName string
		if len(args) == 1 {
			projectName = args[0]
			if !validName.MatchString(projectName) {
				return fmt.Errorf(
					"invalid project name %q: use letters, numbers, hyphens, and underscores only, starting with a letter or number",
					projectName,
				)
			}
		} else {
			var err error
			projectName, err = promptProjectName()
			if err != nil {
				return err
			}
			fmt.Println()
		}

		if !isAllowedTemplate(templateName) {
			return fmt.Errorf(
				"unknown template %q: must be one of %v",
				templateName, allowedTemplates,
			)
		}

		err := project.Create(projectName, templateName, project.Options{
			Verbose: verbose,
		})
		if err != nil {
			return fmt.Errorf("creating project: %w", err)
		}

		fmt.Println()
		fmt.Printf("Project '%s' created successfully.\n", projectName)
		return nil
	},
}

func isAllowedTemplate(name string) bool {
	for _, t := range allowedTemplates {
		if t == name {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().StringVarP(
		&templateName,
		"template",
		"t",
		"fiber",
		fmt.Sprintf("Project template to use (one of: %v)", allowedTemplates),
	)

	newCmd.Flags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"Print each step as it happens (download, extract, organize, customize, install)",
	)
}

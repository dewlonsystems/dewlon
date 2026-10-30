/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dewlonsystems/dewlon/internal/project"
	"github.com/dewlonsystems/dewlon/internal/templates"
	"github.com/spf13/cobra"
)

var (
	templateName string
	targetDir    string
	modulePath   string
	gitRef       string
	skipInstall  bool
)

// promptProjectName keeps asking until it gets a valid name. It takes the
// shared reader so later prompts don't lose buffered input.
func promptProjectName(in *bufio.Reader) (string, error) {
	for {
		fmt.Print("Project name: ")

		line, err := in.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("reading project name: %w", err)
		}

		name := strings.TrimSpace(line)
		if err := project.ValidateName(name); err != nil {
			fmt.Println(err)
			continue
		}
		return name, nil
	}
}

// promptTemplate lists the templates and accepts either a number or a
// name. Pressing enter picks the first one.
func promptTemplate(in *bufio.Reader) (string, error) {
	all := templates.All()

	fmt.Println("Template:")
	for i, t := range all {
		fmt.Printf("  %d) %-8s %s\n", i+1, t.Key, t.Description)
	}

	for {
		fmt.Printf("Choose [1-%d, default 1]: ", len(all))

		line, err := in.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("reading template choice: %w", err)
		}

		choice := strings.TrimSpace(line)
		if choice == "" {
			return all[0].Key, nil
		}
		if n, err := strconv.Atoi(choice); err == nil && n >= 1 && n <= len(all) {
			return all[n-1].Key, nil
		}
		if t, err := templates.Get(choice); err == nil {
			return t.Key, nil
		}
		fmt.Println("Not a valid choice, try again.")
	}
}

var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Create a new Dewlon project",
	Long: `Creates a new project from one of the Dewlon templates.

If project-name is left out and you're in a terminal, you'll be asked for
the name and the template. In scripts and CI, pass the name as an argument.

Examples:

  dewlon new hospital-api
  dewlon new school-api --template=fiber
  dewlon new mobile-app -t flutter
  dewlon new billing --module github.com/acme/billing
  dewlon new clinic-api --dir ~/work --ref v1.2.0
  dewlon new`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: cobra.NoFileCompletions,
	RunE:              runNew,
}

func runNew(cmd *cobra.Command, args []string) error {
	printLogo()

	interactive := isInteractive()
	in := bufio.NewReader(os.Stdin)

	var name string
	if len(args) == 1 {
		name = args[0]
		if err := project.ValidateName(name); err != nil {
			return err
		}
	} else {
		if !interactive {
			return errors.New("no project name given and stdin is not a terminal; pass the name as an argument")
		}
		var err error
		if name, err = promptProjectName(in); err != nil {
			return err
		}
	}

	// Only ask about the template when the person is already being walked
	// through prompts and hasn't picked one with a flag.
	chosen := templateName
	if len(args) == 0 && interactive && !cmd.Flags().Changed("template") {
		var err error
		if chosen, err = promptTemplate(in); err != nil {
			return err
		}
		fmt.Println()
	}

	tpl, err := templates.Get(chosen)
	if err != nil {
		return err
	}

	// Step-by-step lines only appear with --verbose, and the installer's
	// own output is only streamed then too.
	opts := project.Options{
		Dir:         targetDir,
		Module:      modulePath,
		Ref:         gitRef,
		SkipInstall: skipInstall,
	}
	if verbose && !quiet {
		opts.Logf = func(format string, a ...any) { fmt.Printf("  "+format+"\n", a...) }
		opts.InstallOutput = os.Stdout
	}

	say("Creating %s from the %s template...", name, tpl.Key)

	start := time.Now()
	res, err := project.Create(cmd.Context(), name, tpl.Key, opts)
	if err != nil {
		return fmt.Errorf("creating project: %w", err)
	}

	say("")
	success("Created %s in %s", name, time.Since(start).Round(100*time.Millisecond))

	for _, w := range res.Warnings {
		warn("%s", w)
	}

	cd := name
	if targetDir != "" {
		cd = res.Path
	}
	if strings.ContainsAny(cd, " \t") {
		cd = strconv.Quote(cd)
	}

	say("")
	say("Next steps:")
	say("  cd %s", cd)
	if skipInstall && len(tpl.Install) > 0 {
		say("  %s", strings.Join(tpl.Install, " "))
	}
	for _, step := range tpl.NextSteps {
		say("  %s", step)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(newCmd)

	flags := newCmd.Flags()
	flags.StringVarP(&templateName, "template", "t", templates.All()[0].Key,
		fmt.Sprintf("project template to use (%s)", strings.Join(templates.Keys(), ", ")))
	flags.StringVarP(&targetDir, "dir", "d", "", "folder to create the project in (default: current folder)")
	flags.StringVarP(&modulePath, "module", "m", "", "Go module path (default: github.com/dewlonsystems/<name>)")
	flags.StringVar(&gitRef, "ref", "", "template branch, tag or commit to build from (default: the template's main branch)")
	flags.BoolVar(&skipInstall, "no-install", false, "don't install dependencies after creating the project")

	// Tab-complete template names in the shell, with their descriptions.
	_ = newCmd.RegisterFlagCompletionFunc("template",
		func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			var out []string
			for _, t := range templates.All() {
				out = append(out, t.Key+"\t"+t.Description)
			}
			return out, cobra.ShellCompDirectiveNoFileComp
		})
	_ = newCmd.RegisterFlagCompletionFunc("dir",
		func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveFilterDirs
		})
}

// Compile-time check that the verbose installer stream is a plain writer.
var _ io.Writer = os.Stdout

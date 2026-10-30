/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/dewlonsystems/dewlon/internal/buildinfo"
	"github.com/fatih/color"
)

// Small output helpers so every command talks the same way and honours
// --quiet and --no-color without repeating the checks.

// say prints normal output unless --quiet is set.
func say(format string, a ...any) {
	if quiet {
		return
	}
	fmt.Printf(format+"\n", a...)
}

// success prints a green tick line.
func success(format string, a ...any) {
	if quiet {
		return
	}
	fmt.Printf("%s %s\n", color.GreenString("✓"), fmt.Sprintf(format, a...))
}

// warn prints a warning on stderr. Warnings show even with --quiet,
// because they usually need action.
func warn(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "%s %s\n", color.YellowString("warning:"), fmt.Sprintf(format, a...))
}

// hint prints a dim follow-up line on stderr.
func hint(msg string) {
	fmt.Fprintln(os.Stderr, color.New(color.Faint).Sprint(msg))
}

// fail prints an error on stderr. Errors are shown even in quiet mode.
func fail(err error) {
	fmt.Fprintf(os.Stderr, "%s %v\n", color.RedString("error:"), err)
}

// isInteractive reports whether stdin is a real terminal, so we know if
// it makes sense to ask the user questions.
func isInteractive() bool {
	return isTerminal(os.Stdin)
}

// stdoutIsTerminal reports whether stdout is a real terminal.
func stdoutIsTerminal() bool {
	return isTerminal(os.Stdout)
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// printLogo prints the Dewlon mark: a small terminal window holding a
// prompt, with the name and maker beside it.
//
//	┌──────┐
//	│  >_  │  dewlon v0.3.0
//	└──────┘  by Dewlon Systems
//
// Use it sparingly: bare `dewlon`, `dewlon init` and `--version` only.
// Commands that go straight to a prompt (like `new`) should not print it.
// It is skipped with --quiet or when stdout is piped, so scripts and CI
// logs stay clean. Only box-drawing and ASCII characters are used, all
// single-width, so alignment holds in any terminal..
func printLogo() {
	if quiet || !stdoutIsTerminal() {
		return
	}

	info := buildinfo.Get()

	frame := color.New(color.FgCyan).SprintFunc()
	prompt := color.New(color.FgCyan, color.Bold).SprintFunc()
	cursor := color.New(color.FgHiWhite, color.Bold).SprintFunc()
	name := color.New(color.Bold).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()

	fmt.Println()
	fmt.Println("  " + frame("┌──────┐"))
	fmt.Println("  " + frame("│") + "  " + prompt(">") + cursor("_") + "  " + frame("│") + "  " + name("dewlon cli") + " " + dim(info.Version))
	fmt.Println("  " + frame("└──────┘") + "  " + dim("by Dewlon Systems"))
	fmt.Println()
}

// printLogoCompact prints a one-line mark for --version, error screens
// and narrow terminals.
func printLogoCompact() {
	if quiet {
		return
	}
	info := buildinfo.Get()
	fmt.Printf("%s %s %s\n",
		color.New(color.FgCyan, color.Bold).Sprint(">_"),
		color.New(color.Bold).Sprint("dewlon"),
		color.New(color.Faint).Sprint(info.Version),
	)
}

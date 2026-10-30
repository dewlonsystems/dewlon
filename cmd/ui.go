/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import (
	"fmt"
	"os"

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
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// printLogo prints the Dewlon "D" mark with a red accent. Purely cosmetic;
// fatih/color falls back to plain text on terminals without color.
func printLogo() {
	if quiet {
		return
	}
	red := color.New(color.FgRed).SprintFunc()

	fmt.Println(" ██████╗  " + red("███████╗"))
	fmt.Println(" ██╔══██╗ " + red("██╔════╝"))
	fmt.Println(" ██║  ██║ " + red("█████╗  "))
	fmt.Println(" ██║  ██║ " + red("██╔══╝  "))
	fmt.Println(" ██████╔╝ " + red("███████╗"))
	fmt.Println(" ╚═════╝  " + red("╚══════╝"))
	fmt.Println()
}

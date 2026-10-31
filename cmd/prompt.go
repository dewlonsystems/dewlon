/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/dewlonsystems/dewlon/internal/project"
	"github.com/dewlonsystems/dewlon/internal/templates"
)

// readLine reads one line from in, but gives up as soon as ctx is cancelled.
//
// A plain ReadString can't be interrupted: on macOS and Linux a blocked read
// keeps waiting after Ctrl+C, and on Windows the console reports it as
// end-of-input. Reading in the background and waiting on both the read and
// the context handles either case.
//
// End-of-input with nothing typed (Ctrl+D, Ctrl+Z, or Ctrl+C on Windows)
// means the person wants out, so it is reported as a cancellation rather
// than as a failure.
func readLine(ctx context.Context, in *bufio.Reader) (string, error) {
	type result struct {
		line string
		err  error
	}

	done := make(chan result, 1)
	go func() {
		line, err := in.ReadString('\n')
		done <- result{line, err}
	}()

	select {
	case <-ctx.Done():
		// The reading goroutine stays blocked until the process exits,
		// which is what happens right after a cancellation.
		return "", ctx.Err()

	case r := <-done:
		if r.err == nil {
			return r.line, nil
		}
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if errors.Is(r.err, io.EOF) {
			if strings.TrimSpace(r.line) != "" {
				return r.line, nil // last line had no trailing newline
			}
			return "", context.Canceled
		}
		return "", fmt.Errorf("reading input: %w", r.err)
	}
}

// promptProjectName keeps asking until it gets a valid name. It takes the
// shared reader so later prompts don't lose buffered input.
func promptProjectName(ctx context.Context, in *bufio.Reader) (string, error) {
	for {
		fmt.Print("Project name: ")

		line, err := readLine(ctx, in)
		if err != nil {
			return "", err
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
func promptTemplate(ctx context.Context, in *bufio.Reader) (string, error) {
	all := templates.All()

	fmt.Println("Template:")
	for i, t := range all {
		fmt.Printf("  %d) %-8s %s\n", i+1, t.Key, t.Description)
	}

	for {
		fmt.Printf("Choose [1-%d, default 1]: ", len(all))

		line, err := readLine(ctx, in)
		if err != nil {
			return "", err
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

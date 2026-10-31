/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// printTable prints rows under a header with aligned columns.
func printTable(header []string, rows [][]string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	write := func(cols []string) error {
		_, err := fmt.Fprintln(w, strings.Join(cols, "\t"))
		return err
	}

	if err := write(header); err != nil {
		return err
	}
	for _, row := range rows {
		if err := write(row); err != nil {
			return err
		}
	}
	return w.Flush()
}

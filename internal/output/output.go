// SPDX-License-Identifier: Apache-2.0

// Package output renders command results as text or JSON to a writer.
package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSON writes v as indented JSON followed by a newline.
func JSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// Lines writes each string on its own line.
func Lines(w io.Writer, lines []string) error {
	for _, l := range lines {
		if _, err := fmt.Fprintln(w, l); err != nil {
			return err
		}
	}
	return nil
}

// Table writes aligned "Key:  value" rows.
func Table(w io.Writer, rows [][2]string) error {
	width := 0
	for _, r := range rows {
		if len(r[0]) > width {
			width = len(r[0])
		}
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%-*s  %s\n", width+1, r[0]+":", r[1]); err != nil {
			return err
		}
	}
	return nil
}

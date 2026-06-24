// SPDX-License-Identifier: Apache-2.0

package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestJSONIndentsAndEndsWithNewline(t *testing.T) {
	var b bytes.Buffer
	if err := JSON(&b, map[string]int{"count": 2}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(b.String(), "\n") || !strings.Contains(b.String(), "  \"count\": 2") {
		t.Fatalf("unexpected JSON: %q", b.String())
	}
}

func TestLines(t *testing.T) {
	var b bytes.Buffer
	if err := Lines(&b, []string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	if b.String() != "a\nb\n" {
		t.Fatalf("got %q", b.String())
	}
}

func TestTable(t *testing.T) {
	var b bytes.Buffer
	if err := Table(&b, [][2]string{{"Name", "foo"}, {"Version", "1.0"}}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(b.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(lines), lines)
	}

	// Check each line contains its key and value
	if !strings.Contains(lines[0], "Name:") || !strings.Contains(lines[0], "foo") {
		t.Fatalf("line 0 missing key or value: %q", lines[0])
	}
	if !strings.Contains(lines[1], "Version:") || !strings.Contains(lines[1], "1.0") {
		t.Fatalf("line 1 missing key or value: %q", lines[1])
	}

	// Check column alignment: both values start at the same column index
	fooIdx := strings.Index(lines[0], "foo")
	versionValIdx := strings.Index(lines[1], "1.0")
	if fooIdx == -1 || versionValIdx == -1 {
		t.Fatalf("could not find values in output")
	}
	if fooIdx != versionValIdx {
		t.Fatalf("values not column-aligned: foo at %d, 1.0 at %d", fooIdx, versionValIdx)
	}
}

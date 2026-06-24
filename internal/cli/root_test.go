// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"testing"
)

func TestRootHelpListsUseLine(t *testing.T) {
	cmd := New()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help should not error: %v", err)
	}
	if got := out.String(); !bytes.Contains(out.Bytes(), []byte("orca")) {
		t.Fatalf("help output missing tool name; got: %q", got)
	}
}

// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestTagsCommandJSONOutput(t *testing.T) {
	host, craneOpts := startRegistryWithImage(t, "repo", []string{"1.0", "1.1"})
	cmd := New()
	cmd.SetArgs([]string{"tags", host + "/repo", "--json"})
	cmd.SetContext(withTestOCIOptions(context.Background(), craneOpts))
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("tags: %v", err)
	}
	if !strings.Contains(out.String(), "\"count\": 2") {
		t.Fatalf("expected 2 tags in JSON, got: %s", out.String())
	}
}

func TestTagsCommandFilters(t *testing.T) {
	host, craneOpts := startRegistryWithImage(t, "repo", []string{"1.0", "2.0"})
	cmd := New()
	cmd.SetArgs([]string{"tags", host + "/repo", "--filter", "1"})
	cmd.SetContext(withTestOCIOptions(context.Background(), craneOpts))
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("tags: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "1.0") {
		t.Fatalf("expected 1.0 in output, got: %s", output)
	}
	if strings.Contains(output, "2.0") {
		t.Fatalf("expected 2.0 NOT in output, got: %s", output)
	}
}

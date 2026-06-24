// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestCopyDryRunReportsWithoutPushing(t *testing.T) {
	host, opts := startRegistryWithImage(t, "src", []string{"1.0"})
	cmd := New()
	cmd.SetArgs([]string{"copy", host + "/src:1.0", host + "/dst:1.0", "--dry-run"})
	cmd.SetContext(withTestOCIOptions(context.Background(), opts))
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("copy dry-run: %v", err)
	}
	if !strings.Contains(out.String(), "would copy") {
		t.Fatalf("expected dry-run message, got %q", out.String())
	}
}

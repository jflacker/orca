// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestPinCommandPrintsDigestForm(t *testing.T) {
	host, opts := startRegistryWithImage(t, "repo", []string{"1.0"})
	cmd := New()
	cmd.SetArgs([]string{"pin", host + "/repo:1.0"})
	cmd.SetContext(withTestOCIOptions(context.Background(), opts))
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("pin: %v", err)
	}
	if !strings.Contains(out.String(), "@sha256:") {
		t.Fatalf("expected digest form, got %q", out.String())
	}
}

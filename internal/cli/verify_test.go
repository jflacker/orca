// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// TestVerifyCommandRequiresIdentity ensures the command rejects invocations that
// supply neither --certificate-identity nor --certificate-identity-regexp,
// before any network access.
func TestVerifyCommandRequiresIdentity(t *testing.T) {
	cmd := New()
	cmd.SetArgs([]string{"verify", "example.com/repo:1.0"})
	cmd.SetContext(context.Background())
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)
	cmd.SetOut(&errOut)

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "certificate-identity") {
		t.Fatalf("expected identity-required error, got %v", err)
	}
}

// TestVerifyCommandRequiresIssuer ensures that supplying an identity without an
// OIDC issuer fails up front, before any network access. sigstore-go requires
// an issuer whenever a certificate identity is asserted.
func TestVerifyCommandRequiresIssuer(t *testing.T) {
	cmd := New()
	cmd.SetArgs([]string{"verify", "example.com/repo:1.0", "--certificate-identity", "https://example.com/id"})
	cmd.SetContext(context.Background())
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)
	cmd.SetOut(&errOut)

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "certificate-oidc-issuer") {
		t.Fatalf("expected issuer-required error, got %v", err)
	}
}

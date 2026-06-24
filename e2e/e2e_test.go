//go:build e2e

// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os/exec"
	"strings"
	"testing"
)

func TestInspectRealChainguardImage(t *testing.T) {
	out, err := exec.Command("go", "run", "../cmd/orca", "inspect", "cgr.dev/chainguard/static:latest").CombinedOutput() //nolint:gosec // test-only: controlled args, no untrusted input
	if err != nil {
		t.Fatalf("inspect failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Digest") && !strings.Contains(string(out), "image index") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestVerifyRealChainguardImage(t *testing.T) {
	// Exercises the OCI keyless verify path end-to-end against a real signed image,
	// asserting actual cryptographic success. Cosign image signatures are
	// MessageSignatures, so the verifier must supply the signed simple-signing
	// layer's digest as the artifact digest; a regression there surfaces as
	// "artifact must be provided to verify message signature", which we assert
	// against explicitly.
	const realIdentity = "https://github.com/chainguard-images/images/.github/workflows/release.yaml@refs/heads/main"
	const issuer = "https://token.actions.githubusercontent.com"

	out, err := exec.Command( //nolint:gosec // test-only: controlled args, no untrusted input
		"go", "run", "../cmd/orca",
		"verify", "cgr.dev/chainguard/static:latest",
		"--certificate-identity", realIdentity,
		"--certificate-oidc-issuer", issuer,
	).CombinedOutput()

	outStr := string(out)

	if strings.Contains(outStr, "artifact must be provided") {
		t.Fatalf("MessageSignature artifact-digest regression: %s", outStr)
	}
	if err != nil {
		t.Fatalf("verify failed: %v\n%s", err, outStr)
	}
	if !strings.Contains(outStr, "verified") {
		t.Fatalf("verify succeeded but output missing 'verified': %s", outStr)
	}
}

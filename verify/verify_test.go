// SPDX-License-Identifier: Apache-2.0

package verify

import (
	"os"
	"testing"
)

// TestVerifyBundleFileWithRealFixtures verifies a real, committed Sigstore
// bundle (bundle-provenance.json, signed by sigstore-js via GitHub Actions and
// witnessed in Rekor) against the real public-good trusted root. The asserted
// identity is the certificate's actual Fulcio SAN and OIDC issuer, so a passing
// test proves the full signature + transparency-log + trusted-root crypto path.
func TestVerifyBundleFileWithRealFixtures(t *testing.T) {
	trustedRoot, err := os.ReadFile("testdata/trusted-root-public-good.json")
	if err != nil {
		t.Fatalf("read trusted root: %v", err)
	}

	id := Identity{
		OIDCIssuer: "https://token.actions.githubusercontent.com",
		SAN:        "https://github.com/sigstore/sigstore-js/.github/workflows/release.yml@refs/heads/main",
	}

	res, err := BundleFile("testdata/bundle-provenance.json", trustedRoot, id)
	if err != nil {
		t.Fatalf("BundleFile: %v", err)
	}
	if !res.Verified {
		t.Fatalf("expected verified signature, got %+v", res)
	}
	if res.Certificate.Issuer != id.OIDCIssuer {
		t.Errorf("issuer = %q, want %q", res.Certificate.Issuer, id.OIDCIssuer)
	}
	if res.Certificate.SAN != id.SAN {
		t.Errorf("SAN = %q, want %q", res.Certificate.SAN, id.SAN)
	}
}

// TestVerifyBundleFileWrongIdentity ensures the policy is falsifiable: the same
// real bundle must fail when asserting an identity it was not signed with.
func TestVerifyBundleFileWrongIdentity(t *testing.T) {
	trustedRoot, err := os.ReadFile("testdata/trusted-root-public-good.json")
	if err != nil {
		t.Fatalf("read trusted root: %v", err)
	}

	id := Identity{
		OIDCIssuer: "https://token.actions.githubusercontent.com",
		SAN:        "https://github.com/attacker/evil/.github/workflows/release.yml@refs/heads/main",
	}

	_, err = BundleFile("testdata/bundle-provenance.json", trustedRoot, id)
	if err == nil {
		t.Fatalf("expected verification to fail for wrong SAN identity")
	}
}

// TestVerifyMessageSignatureBundle verifies a real cosign image MessageSignature
// offline. cosign-message-signature.json is the bundle assembled from the
// cgr.dev/chainguard/static:latest .sig (Fulcio cert + Rekor entry + the
// simple-signing layer digest). Verifying it offline against the public-good
// trusted root exercises the artifact-digest path that real image signatures
// require — the path that previously failed with "artifact must be provided".
func TestVerifyMessageSignatureBundle(t *testing.T) {
	trustedRoot, err := os.ReadFile("testdata/trusted-root-public-good.json")
	if err != nil {
		t.Fatalf("read trusted root: %v", err)
	}

	id := Identity{
		OIDCIssuer: "https://token.actions.githubusercontent.com",
		SAN:        "https://github.com/chainguard-images/images/.github/workflows/release.yaml@refs/heads/main",
	}

	res, err := MessageSignatureBundleFile("testdata/cosign-message-signature.json", trustedRoot, id)
	if err != nil {
		t.Fatalf("MessageSignatureBundleFile: %v", err)
	}
	if !res.Verified {
		t.Fatalf("expected verified signature, got %+v", res)
	}
	if res.Certificate.Issuer != id.OIDCIssuer {
		t.Errorf("issuer = %q, want %q", res.Certificate.Issuer, id.OIDCIssuer)
	}
	if res.Certificate.SAN != id.SAN {
		t.Errorf("SAN = %q, want %q", res.Certificate.SAN, id.SAN)
	}
}

// TestVerifyMessageSignatureBundleWrongIdentity ensures the message-signature
// policy is falsifiable: the same real signature must fail for a wrong identity.
func TestVerifyMessageSignatureBundleWrongIdentity(t *testing.T) {
	trustedRoot, err := os.ReadFile("testdata/trusted-root-public-good.json")
	if err != nil {
		t.Fatalf("read trusted root: %v", err)
	}

	id := Identity{
		OIDCIssuer: "https://token.actions.githubusercontent.com",
		SAN:        "https://github.com/attacker/evil/.github/workflows/release.yaml@refs/heads/main",
	}

	_, err = MessageSignatureBundleFile("testdata/cosign-message-signature.json", trustedRoot, id)
	if err == nil {
		t.Fatalf("expected verification to fail for wrong SAN identity")
	}
}

// SPDX-License-Identifier: Apache-2.0

// Package verify checks keyless Sigstore (cosign/Fulcio/Rekor) signatures on
// OCI images and on cosign bundle files, mirroring the sigstore-go verification
// examples.
package verify

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/tuf"
	sgverify "github.com/sigstore/sigstore-go/pkg/verify"
)

// Identity constrains the signing certificate. At least one of SAN/SANRegexp
// must be set; OIDCIssuer/OIDCIssuerRegexp are optional but recommended.
type Identity struct {
	OIDCIssuer       string
	OIDCIssuerRegexp string
	SAN              string
	SANRegexp        string
}

// Result reports the verification outcome.
type Result struct {
	Verified    bool   `json:"verified"`
	Digest      string `json:"digest,omitempty"`
	Certificate struct {
		Issuer string `json:"issuer"`
		SAN    string `json:"san"`
	} `json:"certificate"`
}

type config struct {
	trustedRootJSON []byte
	craneOpts       []crane.Option
}

// Option configures Verify.
type Option func(*config)

// WithTrustedRootJSON supplies a pinned trusted root (tests / air-gapped use).
// When absent, Verify fetches the public-good root over TUF.
func WithTrustedRootJSON(b []byte) Option { return func(c *config) { c.trustedRootJSON = b } }

// WithCraneOptions targets a specific registry transport (e.g. tests).
func WithCraneOptions(opts ...crane.Option) Option {
	return func(c *config) { c.craneOpts = append(c.craneOpts, opts...) }
}

// Verify checks the keyless cosign signature on the OCI image ref against the given
// identity, verifying the signature, transparency-log inclusion, and certificate.
func Verify(ctx context.Context, ref string, id Identity, opts ...Option) (*Result, error) {
	var cfg config
	for _, o := range opts {
		o(&cfg)
	}

	material, err := trustedMaterial(ctx, cfg.trustedRootJSON)
	if err != nil {
		return nil, fmt.Errorf("loading trusted material: %w", err)
	}

	b, digestAlg, digestHex, err := bundleFromOCIImage(ctx, ref, cfg.craneOpts...)
	if err != nil {
		return nil, fmt.Errorf("building bundle for %q: %w", ref, err)
	}

	// cosign image signatures are MessageSignatures over the simple-signing payload,
	// whose digest is the .sig layer digest sigstore-go needs as the artifact.
	digestBytes, err := hex.DecodeString(digestHex)
	if err != nil {
		return nil, fmt.Errorf("decoding signed layer digest: %w", err)
	}

	res, err := verifyBundle(b, material, id, sgverify.WithArtifactDigest(digestAlg, digestBytes))
	if err != nil {
		return nil, err
	}
	res.Digest = digestAlg + ":" + digestHex
	return res, nil
}

// BundleFile verifies a cosign/Sigstore bundle stored on disk against the
// supplied trusted root JSON and identity. It is the fully offline, deterministic
// entry point used in tests and for air-gapped verification of captured bundles.
func BundleFile(bundlePath string, trustedRootJSON []byte, id Identity) (*Result, error) {
	if len(trustedRootJSON) == 0 {
		return nil, errors.New("trusted root JSON is required")
	}
	b, err := bundle.LoadJSONFromPath(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("loading bundle %q: %w", bundlePath, err)
	}
	tr, err := root.NewTrustedRootFromJSON(trustedRootJSON)
	if err != nil {
		return nil, fmt.Errorf("parsing trusted root: %w", err)
	}
	material := root.TrustedMaterialCollection{tr}
	// DSSE-envelope bundles carry their artifact in the statement, so none is supplied.
	return verifyBundle(b, material, id, sgverify.WithoutArtifactUnsafe())
}

// MessageSignatureBundleFile verifies an offline cosign MessageSignature bundle
// against the supplied trusted root and identity, using the bundle's own message
// digest as the artifact. It is the offline counterpart to Verify's OCI path.
func MessageSignatureBundleFile(bundlePath string, trustedRootJSON []byte, id Identity) (*Result, error) {
	if len(trustedRootJSON) == 0 {
		return nil, errors.New("trusted root JSON is required")
	}
	b, err := bundle.LoadJSONFromPath(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("loading bundle %q: %w", bundlePath, err)
	}
	sigContent, err := b.SignatureContent()
	if err != nil {
		return nil, fmt.Errorf("reading bundle signature content: %w", err)
	}
	msg := sigContent.MessageSignatureContent()
	if msg == nil {
		return nil, errors.New("bundle does not contain a message signature")
	}
	tr, err := root.NewTrustedRootFromJSON(trustedRootJSON)
	if err != nil {
		return nil, fmt.Errorf("parsing trusted root: %w", err)
	}
	material := root.TrustedMaterialCollection{tr}

	res, err := verifyBundle(b, material, id, sgverify.WithArtifactDigest(msg.DigestAlgorithm(), msg.Digest()))
	if err != nil {
		return nil, err
	}
	res.Digest = msg.DigestAlgorithm() + ":" + hex.EncodeToString(msg.Digest())
	return res, nil
}

// verifyBundle is the shared verification core: it builds a transparency-log +
// observer-timestamp verifier, applies the artifact and certificate-identity
// policies, and maps the verified identity into a Result.
func verifyBundle(b *bundle.Bundle, material root.TrustedMaterial, id Identity, artifactPolicy sgverify.ArtifactPolicyOption) (*Result, error) {
	verifier, err := sgverify.NewVerifier(material,
		sgverify.WithObserverTimestamps(1),
		sgverify.WithTransparencyLog(1),
	)
	if err != nil {
		return nil, fmt.Errorf("building verifier: %w", err)
	}

	certID, err := sgverify.NewShortCertificateIdentity(id.OIDCIssuer, id.OIDCIssuerRegexp, id.SAN, id.SANRegexp)
	if err != nil {
		return nil, fmt.Errorf("building identity policy: %w", err)
	}
	policy := sgverify.NewPolicy(artifactPolicy, sgverify.WithCertificateIdentity(certID))

	vr, err := verifier.Verify(b, policy)
	if err != nil {
		return nil, fmt.Errorf("verification failed: %w", err)
	}

	res := &Result{Verified: true}
	if vr.Signature != nil && vr.Signature.Certificate != nil {
		res.Certificate.Issuer = vr.Signature.Certificate.Issuer
		res.Certificate.SAN = vr.Signature.Certificate.SubjectAlternativeName
	}
	return res, nil
}

// trustedMaterial loads pinned trusted-root JSON when provided, otherwise
// fetches the Sigstore public-good trusted root over TUF.
func trustedMaterial(_ context.Context, pinned []byte) (root.TrustedMaterial, error) {
	if len(pinned) > 0 {
		tr, err := root.NewTrustedRootFromJSON(pinned)
		if err != nil {
			return nil, err
		}
		return root.TrustedMaterialCollection{tr}, nil
	}

	client, err := tuf.New(tuf.DefaultOptions())
	if err != nil {
		return nil, fmt.Errorf("initializing TUF client: %w", err)
	}
	trustedRootJSON, err := client.GetTarget("trusted_root.json")
	if err != nil {
		return nil, fmt.Errorf("fetching trusted_root.json over TUF: %w", err)
	}
	tr, err := root.NewTrustedRootFromJSON(trustedRootJSON)
	if err != nil {
		return nil, fmt.Errorf("parsing TUF trusted root: %w", err)
	}
	return root.TrustedMaterialCollection{tr}, nil
}

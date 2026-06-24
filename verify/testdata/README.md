# verify test fixtures

These are **real** Sigstore artifacts, copied verbatim from the upstream
`github.com/sigstore/sigstore-go@v1.2.1` `examples/` directory. They are not
mocks: the bundles are genuine keyless signatures recorded in the public-good
Rekor transparency log, and `trusted-root-public-good.json` is the real Sigstore
public-good trusted root.

| File | Origin | Notes |
|------|--------|-------|
| `trusted-root-public-good.json` | `examples/trusted-root-public-good.json` | Sigstore public-good Fulcio/Rekor/CT trust anchors. |
| `bundle-provenance.json` | `examples/bundle-provenance.json` | DSSE attestation signed by sigstore-js via GitHub Actions; uses an x509 Fulcio certificate chain. Used by the verification test. |
| `cosign-message-signature.json` | `cgr.dev/chainguard/static:latest` `.sig` | Real cosign image signature assembled into a Sigstore bundle (Fulcio cert + Rekor entry + simple-signing layer digest). Its content is a `MessageSignature`, so verification requires the artifact digest (taken from the bundle's own message digest). Exercises the path real image signatures use. |

## Identity asserted by the test

`bundle-provenance.json` was signed by GitHub Actions OIDC. Decoding its Fulcio
certificate yields:

- **OIDC issuer** (OID 1.3.6.1.4.1.57264.1.8 / .1.1):
  `https://token.actions.githubusercontent.com`
- **SAN** (URI):
  `https://github.com/sigstore/sigstore-js/.github/workflows/release.yml@refs/heads/main`

`TestVerifyBundleFileWithRealFixtures` asserts these exact values, so a pass
proves the full signature + Rekor-inclusion + trusted-root + identity path.

Verification uses the Rekor inclusion / integrated timestamp, so these fixtures
do not expire with the ~10-minute Fulcio certificate lifetime.

`cosign-message-signature.json` is signed by the Chainguard images GitHub Actions
workflow. Decoding its Fulcio certificate yields:

- **OIDC issuer**: `https://token.actions.githubusercontent.com`
- **SAN** (URI):
  `https://github.com/chainguard-images/images/.github/workflows/release.yaml@refs/heads/main`

`TestVerifyMessageSignatureBundle` asserts these exact values.

## Regenerating

Re-copy from a newer sigstore-go release:

```sh
MC="$(go env GOMODCACHE)/github.com/sigstore/sigstore-go@vX.Y.Z"
cp "$MC/examples/trusted-root-public-good.json" .
cp "$MC/examples/bundle-provenance.json" .
```

Then re-derive the certificate identity for the new `bundle-provenance.json`
(decode `verificationMaterial.x509CertificateChain.certificates[0].rawBytes`
with `openssl x509 -text`) and update the asserted identity in
`verify/verify_test.go`.

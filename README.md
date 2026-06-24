# orca

`orca` is a focused CLI for working with OCI registries: listing tags, inspecting manifests, pinning mutable tags to immutable digests, mirroring images between registries, and verifying keyless Sigstore signatures. It is intentionally narrow — each command does one thing and exits cleanly.

## Install

```
go install github.com/jflacker/orca/cmd/orca@latest
```

## Usage

### tags

List tags for a repository.

```
orca tags cgr.dev/chainguard/static
orca tags cgr.dev/chainguard/go --filter latest --limit 10
orca tags docker.io/library/alpine --json
```

Flags: `--filter <substring>`, `--limit <n>` (0 = all), `--json`

---

### inspect

Show manifest, config, and size for an image or index. When given a multi-arch index with no flags, prints an index summary (digest + platform list) rather than silently resolving to the host platform.

```
orca inspect cgr.dev/chainguard/static:latest
orca inspect cgr.dev/chainguard/go:latest --platform linux/arm64
orca inspect cgr.dev/chainguard/go:latest --all-platforms
orca inspect docker.io/library/alpine:latest --json
```

Flags: `--platform <os/arch[/variant]>`, `--all-platforms`, `--json`
(`--platform` and `--all-platforms` are mutually exclusive)

---

### pin

Resolve a mutable tag to its immutable digest reference (`repo@sha256:...`).

```
orca pin cgr.dev/chainguard/static:latest
orca pin docker.io/library/alpine:3 --json
```

Flags: `--json`

---

### copy

Mirror an image (all artifacts, including attestations and index children) from one registry to another. `--dry-run` resolves and reports the digest without pushing.

```
orca copy cgr.dev/chainguard/static:latest ttl.sh/my-static:latest
orca copy docker.io/library/alpine:3 registry.example.com/mirror/alpine:3 --dry-run
```

Flags: `--dry-run`

---

### verify

Verify a keyless cosign signature (Fulcio certificate + Rekor log) against a pinned trusted root. At least one of `--certificate-identity` or `--certificate-identity-regexp` is required.

```
orca verify cgr.dev/chainguard/static:latest \
  --certificate-identity https://github.com/chainguard-images/images/.github/workflows/release.yaml@refs/heads/main \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com

orca verify docker.io/library/alpine:latest \
  --certificate-identity-regexp ".*alpine.*" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --json
```

Flags: `--certificate-identity <san>`, `--certificate-identity-regexp <regexp>`, `--certificate-oidc-issuer <url>`, `--json`

---

## Design notes

`orca` is built on `go-containerregistry`'s `crane` abstraction for all standard registry operations (pull, push, tag listing, manifest fetch, image copy). It drops down to the lower-level `remote` package for index introspection generally — it backs `inspect`, `pin`, the `IsIndex`/`IndexSummary` helpers, and `copy`'s source resolution (via `get`) — wherever `crane` lacks the API. Bounded concurrency for per-platform fetches lives in a single site in the OCI client, using an `errgroup` from `golang.org/x/sync/errgroup` with a concurrency limit set via `errgroup.SetLimit`; each goroutine writes only its own result slot, avoiding race conditions on per-index result collection.

`inspect` defaults to an honest index summary when given a multi-arch reference: it reports the digest and platform list rather than silently selecting the host platform. Explicit platform selection requires `--platform`.

`verify` uses `sigstore-go` directly against a pinned trusted root (with Rekor inclusion time checked), mirroring the sigstore-go API example rather than shelling out to the cosign binary. Releases of `orca` itself are keyless-cosign-signed via GoReleaser — the `.sig` and `.pem` files shipped with each release can be verified with `cosign verify-blob`.

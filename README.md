# orca

[![CI](https://github.com/jflacker/orca/actions/workflows/ci.yml/badge.svg)](https://github.com/jflacker/orca/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/jflacker/orca.svg)](https://pkg.go.dev/github.com/jflacker/orca)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

> A focused Go CLI and library for OCI registries — list tags, inspect manifests, pin tags to digests, mirror images, and verify keyless Sigstore signatures.

`orca` is intentionally narrow: each command does one thing, talks to the registry over a shared retrying transport, and exits cleanly. It's built directly on the libraries the cloud-native ecosystem standardizes on — [`go-containerregistry`](https://github.com/google/go-containerregistry) for registry access and [`sigstore-go`](https://github.com/sigstore/sigstore-go) for keyless signature verification — rather than shelling out to other CLIs.

| Command | What it does |
|---------|--------------|
| `orca tags <repo>` | List the tags in a repository (filter, limit, JSON) |
| `orca inspect <ref>` | Show manifest/config/size, or an honest multi-arch index summary |
| `orca pin <ref>` | Resolve a mutable tag to its immutable `repo@sha256:…` digest |
| `orca copy <src> <dst>` | Mirror an image between registries (with `--dry-run`) |
| `orca verify <ref>` | Verify a keyless cosign/Fulcio/Rekor signature |

## Install

```
go install github.com/jflacker/orca/cmd/orca@latest
```

Or grab a signed binary from the [releases page](https://github.com/jflacker/orca/releases) — each archive ships an SBOM and a keyless cosign signature (see [Design notes](#design-notes)).

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

Verify a keyless cosign signature (Fulcio certificate + Rekor log) against the Sigstore public-good trusted root. At least one of `--certificate-identity` / `--certificate-identity-regexp` is required, along with `--certificate-oidc-issuer`.

```
orca verify cgr.dev/chainguard/static:latest \
  --certificate-identity https://github.com/chainguard-images/images/.github/workflows/release.yaml@refs/heads/main \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

Flags: `--certificate-identity <san>`, `--certificate-identity-regexp <regexp>`, `--certificate-oidc-issuer <url>`, `--json`

---

## Why I built this

Most everyday registry tasks — checking which tags exist, seeing an image's real digest and size, pinning a tag for a reproducible build, or confirming an image is actually signed — don't need a full `docker pull`. I wanted one small, fast, honest tool for them, built straight on the ecosystem's own libraries instead of wrapping other binaries.

It also doubles as a compact, readable reference for those libraries — in particular the cosign → Sigstore-bundle verification path (reconstructing a Sigstore bundle from a cosign OCI signature's annotations and verifying it with `sigstore-go`), which isn't documented in many places.

## Design notes

`orca` is built on `go-containerregistry`'s `crane` porcelain for standard registry operations (tag listing, manifest/config fetch, image copy). It drops to the lower-level `remote` package only for index introspection — so `inspect`, `pin`, and `copy` can branch on image-vs-index explicitly instead of letting `crane` silently resolve a multi-arch index to the host platform.

Concurrency lives in exactly one place: `inspect --all-platforms` fetches each platform's config in parallel via `errgroup` with a bounded `SetLimit`. Each goroutine writes only its own result slot, so the shared slice needs no lock (verified under `go test -race`); a single platform's failure is recorded, not fatal.

`verify` uses `sigstore-go` directly against the public-good trusted root (with Rekor inclusion time honored), mirroring the upstream API rather than shelling out to the cosign binary. Fittingly, `orca`'s own releases are keyless-cosign-signed via GoReleaser — every release archive ships a `.sig` + `.pem` you can check with `cosign verify-blob`, plus an SBOM.

Tests run against a real in-memory registry (`go-containerregistry`'s `registry` package) and real committed Sigstore fixtures — no mocks.

## License

[Apache-2.0](LICENSE).

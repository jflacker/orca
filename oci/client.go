// SPDX-License-Identifier: Apache-2.0

// Package oci provides a small, context-aware client for querying OCI
// registries. It is built on go-containerregistry's high-level crane porcelain
// and drops to the remote package only for index introspection.
package oci

import (
	"context"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

const defaultConcurrency = 8

// Client performs OCI registry operations with a shared, retrying transport.
type Client struct {
	craneOptions []crane.Option
	transport    http.RoundTripper
	insecure     bool
	limit        int
}

// Option configures a Client.
type Option func(*Client)

// WithCraneOptions appends raw crane options, applied after the client's defaults
// (so a crane.WithTransport here overrides the retrying transport).
func WithCraneOptions(opts ...crane.Option) Option {
	return func(c *Client) { c.craneOptions = append(c.craneOptions, opts...) }
}

// WithTransport sets the RoundTripper for registry calls. It is wrapped with
// retry/backoff (honoring Retry-After); the default keychain still applies.
func WithTransport(rt http.RoundTripper) Option {
	return func(c *Client) { c.transport = rt }
}

// WithInsecure allows plain-HTTP registries; intended for tests against a local registry.
func WithInsecure() Option {
	return func(c *Client) { c.insecure = true }
}

// WithConcurrency sets the bound for parallel per-platform fetches.
func WithConcurrency(n int) Option {
	return func(c *Client) { c.limit = n }
}

// NewClient builds a Client. By default it uses the ambient keychain and a
// retrying transport that honors Retry-After on 429/5xx.
func NewClient(opts ...Option) *Client {
	c := &Client{}
	for _, o := range opts {
		o(c)
	}
	return c
}

func (c *Client) concurrency() int {
	if c.limit <= 0 {
		return defaultConcurrency
	}
	return c.limit
}

func (c *Client) craneOpts(ctx context.Context) []crane.Option {
	base := []crane.Option{
		crane.WithContext(ctx),
		crane.WithAuthFromKeychain(authn.DefaultKeychain),
		crane.WithTransport(transport.NewRetry(c.roundTripper())),
	}
	if c.insecure {
		base = append(base, crane.Insecure)
	}
	return append(base, c.craneOptions...)
}

// roundTripper returns the client's transport, falling back to http.DefaultTransport.
func (c *Client) roundTripper() http.RoundTripper {
	if c.transport != nil {
		return c.transport
	}
	return http.DefaultTransport
}

// remoteOpts builds remote.Options from the client's auth/transport configuration.
func (c *Client) remoteOpts(ctx context.Context) []remote.Option {
	return []remote.Option{
		remote.WithContext(ctx),
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
		remote.WithTransport(transport.NewRetry(c.roundTripper())),
	}
}

// parseRef parses a reference, applying name.Insecure when the client is in insecure mode.
func (c *Client) parseRef(ref string) (name.Reference, error) {
	if c.insecure {
		return name.ParseReference(ref, name.Insecure)
	}
	return name.ParseReference(ref)
}

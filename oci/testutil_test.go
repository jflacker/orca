// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// newTestRegistry starts a real in-memory OCI registry and returns the host,
// crane options that target it over plain HTTP, and the underlying test server.
func newTestRegistry(t *testing.T) (string, []crane.Option, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(registry.New())
	t.Cleanup(srv.Close)
	host := strings.TrimPrefix(srv.URL, "http://")
	opts := []crane.Option{
		crane.WithTransport(srv.Client().Transport),
		crane.Insecure,
	}
	return host, opts, srv
}

func pushImage(t *testing.T, ref string, opts ...crane.Option) v1.Hash {
	t.Helper()
	img, err := random.Image(1024, 3)
	if err != nil {
		t.Fatalf("random.Image: %v", err)
	}
	if err := crane.Push(img, ref, opts...); err != nil {
		t.Fatalf("push %s: %v", ref, err)
	}
	d, err := img.Digest()
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	return d
}

func pushIndex(t *testing.T, ref string, platforms []string, transport http.RoundTripper) {
	t.Helper()
	idx := v1.ImageIndex(empty.Index)
	for _, p := range platforms {
		img, err := random.Image(1024, 2)
		if err != nil {
			t.Fatalf("random.Image: %v", err)
		}
		parts := strings.SplitN(p, "/", 2)
		idx = mutate.AppendManifests(idx, mutate.IndexAddendum{
			Add:        img,
			Descriptor: v1.Descriptor{Platform: &v1.Platform{OS: parts[0], Architecture: parts[1]}},
		})
	}
	r, err := name.ParseReference(ref, name.Insecure)
	if err != nil {
		t.Fatalf("parse %s: %v", ref, err)
	}
	if err := remote.WriteIndex(r, idx, remote.WithTransport(transport)); err != nil {
		t.Fatalf("write index %s: %v", ref, err)
	}
}

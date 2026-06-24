// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func TestInspectAllPlatformsReturnsOnePerPlatform(t *testing.T) {
	host, _, srv := newTestRegistry(t)
	pushIndex(t, host+"/idx:1.0", []string{"linux/amd64", "linux/arm64", "linux/arm/v7"}, srv.Client().Transport)

	c := NewClient(WithInsecure(), WithTransport(srv.Client().Transport), WithConcurrency(2))
	results, err := c.InspectAllPlatforms(context.Background(), host+"/idx:1.0")
	if err != nil {
		t.Fatalf("InspectAllPlatforms: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("want 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("platform %s failed: %v", r.Platform, r.Err)
		}
		if r.Info == nil || r.Info.Digest == "" {
			t.Fatalf("platform %s missing info", r.Platform)
		}
	}
}

// TestInspectAllPlatformsPartialFailureKeepsGoing verifies that when one
// platform's manifest fetch returns a real HTTP 500, InspectAllPlatforms
// returns (results, nil) with that platform's Err populated and siblings
// unaffected.
func TestInspectAllPlatformsPartialFailureKeepsGoing(t *testing.T) {
	// Build a real in-memory registry and push a 2-platform index into it.
	innerHandler := registry.New()
	pushSrv := httptest.NewServer(innerHandler)
	t.Cleanup(pushSrv.Close)
	innerHost := strings.TrimPrefix(pushSrv.URL, "http://")

	pushIndex(t, innerHost+"/idx:1.0", []string{"linux/amd64", "linux/arm64"}, pushSrv.Client().Transport)

	// Fetch the index manifest to discover the arm64 child digest.
	idxRef, err := name.ParseReference(innerHost+"/idx:1.0", name.Insecure)
	if err != nil {
		t.Fatalf("parse ref: %v", err)
	}
	desc, err := remote.Get(idxRef, remote.WithTransport(pushSrv.Client().Transport))
	if err != nil {
		t.Fatalf("remote.Get index: %v", err)
	}
	idx, err := desc.ImageIndex()
	if err != nil {
		t.Fatalf("ImageIndex: %v", err)
	}
	idxManifest, err := idx.IndexManifest()
	if err != nil {
		t.Fatalf("IndexManifest: %v", err)
	}

	// Identify the arm64 child digest to inject failures for.
	var brokenDigest string
	for _, m := range idxManifest.Manifests {
		if m.Platform != nil && m.Platform.Architecture == "arm64" {
			brokenDigest = m.Digest.String()
			break
		}
	}
	if brokenDigest == "" {
		t.Fatal("arm64 descriptor not found in index manifest")
	}

	// Wrap the inner handler: return 500 for manifest GETs matching the
	// arm64 child digest; delegate everything else to the real registry.
	faultyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, brokenDigest) {
			http.Error(w, "injected failure", http.StatusInternalServerError)
			return
		}
		innerHandler.ServeHTTP(w, r)
	})
	testSrv := httptest.NewServer(faultyHandler)
	t.Cleanup(testSrv.Close)
	testHost := strings.TrimPrefix(testSrv.URL, "http://")

	c := NewClient(WithInsecure(), WithTransport(testSrv.Client().Transport), WithConcurrency(2))
	results, err := c.InspectAllPlatforms(context.Background(), testHost+"/idx:1.0")
	if err != nil {
		t.Fatalf("InspectAllPlatforms returned error, want nil: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 results, got %d", len(results))
	}

	var failed, succeeded int
	for _, r := range results {
		if r.Err != nil {
			failed++
			if r.Info != nil {
				t.Errorf("platform %s: Err set but Info non-nil", r.Platform)
			}
		} else {
			succeeded++
			if r.Info == nil || r.Info.Digest == "" {
				t.Errorf("platform %s: no error but Info missing digest", r.Platform)
			}
		}
	}
	if failed != 1 {
		t.Errorf("want exactly 1 failed platform, got %d", failed)
	}
	if succeeded < 1 {
		t.Errorf("want at least 1 succeeded platform, got %d", succeeded)
	}
}

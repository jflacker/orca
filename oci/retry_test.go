// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/registry"
)

func TestListTagsRetriesOn429(t *testing.T) {
	var hits int32
	inner := registry.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/tags/list") && atomic.AddInt32(&hits, 1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		inner.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	host := strings.TrimPrefix(srv.URL, "http://")
	pushImage(t, host+"/repo:1.0", crane.WithTransport(srv.Client().Transport), crane.Insecure)

	c := NewClient(WithTransport(srv.Client().Transport), WithCraneOptions(crane.Insecure))
	tags, err := c.ListTags(context.Background(), host+"/repo")
	if err != nil {
		t.Fatalf("ListTags should recover from 429: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("want 1 tag after retry, got %v", tags)
	}
	if atomic.LoadInt32(&hits) < 2 {
		t.Fatalf("expected a retry after 429, hits=%d", hits)
	}
}

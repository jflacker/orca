// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"
	"strings"
	"testing"
)

func TestPinReturnsIndexDigestForIndex(t *testing.T) {
	host, _, srv := newTestRegistry(t)
	pushIndex(t, host+"/idx:1.0", []string{"linux/amd64", "linux/arm64"}, srv.Client().Transport)

	c := NewClient(WithInsecure(), WithTransport(srv.Client().Transport))
	pinned, err := c.Pin(context.Background(), host+"/idx:1.0")
	if err != nil {
		t.Fatalf("Pin: %v", err)
	}
	sum, err := c.IndexSummary(context.Background(), host+"/idx:1.0")
	if err != nil {
		t.Fatalf("IndexSummary: %v", err)
	}
	if !strings.HasSuffix(pinned, "@"+sum.Digest) {
		t.Fatalf("pin %q should end with index digest %q", pinned, sum.Digest)
	}
	if !strings.HasPrefix(pinned, host+"/idx@") {
		t.Fatalf("pin %q should keep repository path", pinned)
	}
}

// SPDX-License-Identifier: Apache-2.0

package oci

import "testing"

func TestClientConcurrencyDefaultAndOverride(t *testing.T) {
	if got := NewClient().concurrency(); got != 8 {
		t.Fatalf("default concurrency = %d, want 8", got)
	}
	if got := NewClient(WithConcurrency(3)).concurrency(); got != 3 {
		t.Fatalf("override concurrency = %d, want 3", got)
	}
	if got := NewClient(WithConcurrency(0)).concurrency(); got != 8 {
		t.Fatalf("zero concurrency should fall back to 8, got %d", got)
	}
}

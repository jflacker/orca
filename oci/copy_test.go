// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"
	"testing"
)

func TestCopyMovesImageBetweenRepos(t *testing.T) {
	host, craneOpts, srv := newTestRegistry(t)
	pushImage(t, host+"/src:1.0", craneOpts...)

	c := NewClient(WithInsecure(), WithTransport(srv.Client().Transport))
	ctx := context.Background()

	// Dry run must NOT create the destination.
	digest, err := c.Copy(ctx, host+"/src:1.0", host+"/dst:1.0", CopyOptions{DryRun: true})
	if err != nil {
		t.Fatalf("dry-run copy: %v", err)
	}
	if digest == "" {
		t.Fatal("dry-run copy: expected non-empty digest")
	}
	if _, err := c.ListTags(ctx, host+"/dst"); err == nil {
		t.Fatal("dry-run should not have created destination repo")
	}

	// Real copy creates it.
	if _, err := c.Copy(ctx, host+"/src:1.0", host+"/dst:1.0", CopyOptions{}); err != nil {
		t.Fatalf("copy: %v", err)
	}
	tags, err := c.ListTags(ctx, host+"/dst")
	if err != nil || len(tags) != 1 {
		t.Fatalf("destination tags = %v, err = %v", tags, err)
	}

	// Idempotent re-copy.
	if _, err := c.Copy(ctx, host+"/src:1.0", host+"/dst:1.0", CopyOptions{}); err != nil {
		t.Fatalf("re-copy should be idempotent: %v", err)
	}
}

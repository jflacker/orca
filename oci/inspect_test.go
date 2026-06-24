// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"
	"testing"
)

func TestIsIndexTrueForIndexFalseForImage(t *testing.T) {
	host, craneOpts, srv := newTestRegistry(t)
	pushImage(t, host+"/img:1.0", craneOpts...)
	pushIndex(t, host+"/idx:1.0", []string{"linux/amd64", "linux/arm64"}, srv.Client().Transport)

	c := NewClient(WithInsecure(), WithTransport(srv.Client().Transport))
	ctx := context.Background()

	if ok, err := c.IsIndex(ctx, host+"/img:1.0"); err != nil || ok {
		t.Fatalf("image: ok=%v err=%v, want false,nil", ok, err)
	}
	if ok, err := c.IsIndex(ctx, host+"/idx:1.0"); err != nil || !ok {
		t.Fatalf("index: ok=%v err=%v, want true,nil", ok, err)
	}
}

func TestIndexSummaryListsPlatforms(t *testing.T) {
	host, craneOpts, srv := newTestRegistry(t)
	_ = craneOpts
	pushIndex(t, host+"/idx:1.0", []string{"linux/amd64", "linux/arm64"}, srv.Client().Transport)

	c := NewClient(WithInsecure(), WithTransport(srv.Client().Transport))
	sum, err := c.IndexSummary(context.Background(), host+"/idx:1.0")
	if err != nil {
		t.Fatalf("IndexSummary: %v", err)
	}
	if len(sum.Platforms) != 2 {
		t.Fatalf("want 2 platforms, got %v", sum.Platforms)
	}
}

func TestInspectImageReportsDigestAndLayers(t *testing.T) {
	host, craneOpts, srv := newTestRegistry(t)
	pushImage(t, host+"/img:1.0", craneOpts...)

	c := NewClient(WithInsecure(), WithTransport(srv.Client().Transport))
	info, err := c.InspectImage(context.Background(), host+"/img:1.0", nil)
	if err != nil {
		t.Fatalf("InspectImage: %v", err)
	}
	if info.Digest == "" || info.Layers == 0 {
		t.Fatalf("incomplete info: %+v", info)
	}
}

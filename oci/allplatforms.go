// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"golang.org/x/sync/errgroup"
)

// PlatformResult is the inspection outcome for one platform of an index.
type PlatformResult struct {
	Platform Platform
	Info     *ImageInfo
	Err      error
}

// InspectAllPlatforms inspects every platform of an index concurrently with a
// bounded pool. Each goroutine writes only its own slot (no lock); a per-platform
// failure is recorded in PlatformResult.Err. It errors only on cancellation or if
// every platform failed.
func (c *Client) InspectAllPlatforms(ctx context.Context, ref string) ([]PlatformResult, error) {
	sum, err := c.IndexSummary(ctx, ref)
	if err != nil {
		return nil, err
	}
	results := make([]PlatformResult, len(sum.Platforms))

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(c.concurrency())
	for i, p := range sum.Platforms {
		i, p := i, p
		results[i].Platform = p
		g.Go(func() error {
			plat := &v1.Platform{OS: p.OS, Architecture: p.Architecture, Variant: p.Variant}
			info, err := c.InspectImage(gctx, ref, plat)
			// record the error; returning nil keeps siblings running
			results[i].Info, results[i].Err = info, err
			return nil
		})
	}
	_ = g.Wait() // always nil; goroutines never return an error

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var firstErr error
	allFailed := len(results) > 0
	for _, r := range results {
		if r.Err == nil {
			allFailed = false
		} else if firstErr == nil {
			firstErr = r.Err
		}
	}
	if allFailed {
		return results, fmt.Errorf("all %d platforms failed: %w", len(results), firstErr)
	}
	return results, nil
}

// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/crane"
)

// CopyOptions controls Copy behavior.
type CopyOptions struct {
	DryRun bool
}

// Copy mirrors src to dst faithfully — for a multi-arch index, all platforms are
// copied by default because crane.Copy preserves the full artifact. It validates dst
// first (fail fast), then either reports the source digest (DryRun) or performs
// crane.Copy. Copy is idempotent by digest.
func (c *Client) Copy(ctx context.Context, src, dst string, opts CopyOptions) (string, error) {
	// Fail fast: a malformed destination shouldn't cost a source pull.
	if _, err := c.parseRef(dst); err != nil {
		return "", fmt.Errorf("parsing destination %q: %w", dst, err)
	}
	desc, err := c.get(ctx, src)
	if err != nil {
		return "", fmt.Errorf("resolving source %q: %w", src, err)
	}
	if opts.DryRun {
		return desc.Digest.String(), nil
	}
	if err := crane.Copy(src, dst, c.craneOpts(ctx)...); err != nil {
		return "", fmt.Errorf("copying %q to %q: %w", src, dst, err)
	}
	return desc.Digest.String(), nil
}

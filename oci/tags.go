// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
)

// ListTags returns all tags for a repository.
func (c *Client) ListTags(ctx context.Context, repo string) ([]string, error) {
	tags, err := crane.ListTags(repo, c.craneOpts(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("listing tags for %q: %w", repo, err)
	}
	return tags, nil
}

// FilterAndLimit applies a substring filter, then truncates to limit (0 = all).
// Filter is always applied before limit so callers never resolve more than asked.
func FilterAndLimit(tags []string, filter string, limit int) []string {
	out := tags
	if filter != "" {
		out = out[:0:0]
		for _, t := range tags {
			if strings.Contains(t, filter) {
				out = append(out, t)
			}
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

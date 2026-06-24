// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"
	"fmt"
)

// Pin resolves ref to its immutable digest form repo@sha256:.... For an index,
// it pins the index digest itself rather than any single platform.
func (c *Client) Pin(ctx context.Context, ref string) (string, error) {
	desc, err := c.get(ctx, ref)
	if err != nil {
		return "", err
	}
	r, err := c.parseRef(ref)
	if err != nil {
		return "", fmt.Errorf("parsing reference %q: %w", ref, err)
	}
	return r.Context().Name() + "@" + desc.Digest.String(), nil
}

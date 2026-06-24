// SPDX-License-Identifier: Apache-2.0

// Package cli implements the orca command-line interface.
package cli

import (
	"context"

	"github.com/jflacker/orca/oci"
)

// ctxKey is an unexported context key type for injecting test oci options.
type ctxKey struct{}

// withOCIOptions injects oci.Client options (tests only).
func withOCIOptions(ctx context.Context, opts []oci.Option) context.Context {
	return context.WithValue(ctx, ctxKey{}, opts)
}

// clientFromContext returns an oci.Client, applying any injected test options.
func clientFromContext(ctx context.Context) *oci.Client {
	if opts, ok := ctx.Value(ctxKey{}).([]oci.Option); ok && len(opts) > 0 {
		return oci.NewClient(opts...)
	}
	return oci.NewClient()
}

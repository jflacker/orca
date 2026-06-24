// SPDX-License-Identifier: Apache-2.0

// Command orca is a CLI for querying and verifying container images in OCI registries.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jflacker/orca/internal/cli"
)

// version is the build version, overridden at release time via
// -ldflags "-X main.version=...". GoReleaser sets this automatically.
var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cmd := cli.New()
	cmd.Version = version

	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

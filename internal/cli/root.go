// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"log/slog"
	"os"

	"github.com/chainguard-dev/clog"
	"github.com/spf13/cobra"
)

// New builds the root orca command with all subcommands attached.
func New() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:           "orca",
		Short:         "Query and verify container images in OCI registries",
		Long:          "orca is a CLI for exploring and verifying OCI registry images: list tags, inspect manifests, pin tags to digests, copy images, and verify keyless Sigstore signatures.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			level := slog.LevelInfo
			if verbose {
				level = slog.LevelDebug
			}
			h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
			log := clog.New(h)
			cmd.SetContext(clog.WithLogger(cmd.Context(), log))
			return nil
		},
	}
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging to stderr")

	cmd.AddCommand(tagsCmd())
	cmd.AddCommand(inspectCmd())
	cmd.AddCommand(pinCmd())
	cmd.AddCommand(copyCmd())
	cmd.AddCommand(verifyCmd())
	return cmd
}

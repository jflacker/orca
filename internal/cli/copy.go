// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"

	"github.com/jflacker/orca/oci"
	"github.com/spf13/cobra"
)

func copyCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "copy <src> <dst>",
		Short: "Mirror an image between registries",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			src, dst := args[0], args[1]

			digest, err := clientFromContext(ctx).Copy(ctx, src, dst, oci.CopyOptions{DryRun: dryRun})
			if err != nil {
				return err
			}
			if dryRun {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "would copy %s -> %s (%s)\n", src, dst, digest)
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "copied %s -> %s (%s)\n", src, dst, digest)
			return err
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Resolve and report without pushing")
	return cmd
}

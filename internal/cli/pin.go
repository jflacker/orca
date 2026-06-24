// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"

	"github.com/jflacker/orca/internal/output"
	"github.com/spf13/cobra"
)

func pinCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "pin <reference>",
		Short: "Resolve a tag to its immutable digest (repo@sha256:...)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			pinned, err := clientFromContext(ctx).Pin(ctx, args[0])
			if err != nil {
				return err
			}
			if asJSON {
				return output.JSON(cmd.OutOrStdout(), struct {
					Reference string `json:"reference"`
					Pinned    string `json:"pinned"`
				}{args[0], pinned})
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), pinned)
			return err
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

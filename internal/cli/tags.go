// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/chainguard-dev/clog"
	"github.com/jflacker/orca/internal/output"
	"github.com/jflacker/orca/oci"
	"github.com/spf13/cobra"
)

func tagsCmd() *cobra.Command {
	var (
		filter string
		limit  int
		asJSON bool
	)
	cmd := &cobra.Command{
		Use:   "tags <repository>",
		Short: "List tags for a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := clog.FromContext(ctx)
			repo := args[0]

			log.Debugf("listing tags for %s", repo)
			client := clientFromContext(ctx)
			all, err := client.ListTags(ctx, repo)
			if err != nil {
				return err
			}
			tags := oci.FilterAndLimit(all, filter, limit)

			if asJSON {
				return output.JSON(cmd.OutOrStdout(), struct {
					Repository string   `json:"repository"`
					Count      int      `json:"count"`
					Tags       []string `json:"tags"`
				}{repo, len(tags), tags})
			}
			return output.Lines(cmd.OutOrStdout(), tags)
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "Only show tags containing this substring")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum tags to show (0 = all)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"io"
	"strconv"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/jflacker/orca/internal/output"
	"github.com/jflacker/orca/oci"
	"github.com/spf13/cobra"
)

func inspectCmd() *cobra.Command {
	var (
		platform     string
		allPlatforms bool
		asJSON       bool
	)
	cmd := &cobra.Command{
		Use:   "inspect <reference>",
		Short: "Show manifest, config, and size for an image or index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ref := args[0]
			client := clientFromContext(ctx)
			w := cmd.OutOrStdout()

			if allPlatforms {
				results, err := client.InspectAllPlatforms(ctx, ref)
				if err != nil {
					return err
				}
				return renderAllPlatforms(cmd.OutOrStdout(), results, asJSON)
			}

			if platform == "" {
				isIdx, err := client.IsIndex(ctx, ref)
				if err != nil {
					return err
				}
				if isIdx {
					sum, err := client.IndexSummary(ctx, ref)
					if err != nil {
						return err
					}
					return renderIndexSummary(w, sum, asJSON)
				}
			}

			pl, err := platformOrNil(platform)
			if err != nil {
				return err
			}
			info, err := client.InspectImage(ctx, ref, pl)
			if err != nil {
				return err
			}
			return renderImage(w, info, asJSON)
		},
	}
	cmd.Flags().StringVar(&platform, "platform", "", "Inspect a specific platform (os/arch[/variant])")
	cmd.Flags().BoolVar(&allPlatforms, "all-platforms", false, "Inspect every platform of an index")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	cmd.MarkFlagsMutuallyExclusive("platform", "all-platforms")
	return cmd
}

func platformOrNil(s string) (*v1.Platform, error) {
	if s == "" {
		return nil, nil
	}
	return parsePlatform(s)
}

func renderImage(w io.Writer, info *oci.ImageInfo, asJSON bool) error {
	if asJSON {
		return output.JSON(w, info)
	}
	return output.Table(w, [][2]string{
		{"Reference", info.Reference},
		{"Type", "image"},
		{"Digest", info.Digest},
		{"Media Type", info.MediaType},
		{"Platform", info.OS + "/" + info.Architecture},
		{"Created", info.Created.String()},
		{"Layers", strconv.Itoa(info.Layers)},
		{"Size", formatBytes(info.Size)},
	})
}

func renderIndexSummary(w io.Writer, sum *oci.IndexSummary, asJSON bool) error {
	if asJSON {
		return output.JSON(w, sum)
	}
	names := make([]string, len(sum.Platforms))
	for i, p := range sum.Platforms {
		names[i] = p.String()
	}
	return output.Table(w, [][2]string{
		{"Reference", sum.Reference},
		{"Type", "image index (multi-arch)"},
		{"Digest", sum.Digest},
		{"Media Type", sum.MediaType},
		{"Platforms", strings.Join(names, ", ")},
	})
}

type platformResultJSON struct {
	Platform string         `json:"platform"`
	Info     *oci.ImageInfo `json:"info,omitempty"`
	Error    string         `json:"error,omitempty"`
}

func renderAllPlatforms(w io.Writer, results []oci.PlatformResult, asJSON bool) error {
	if asJSON {
		out := make([]platformResultJSON, len(results))
		for i, r := range results {
			out[i] = platformResultJSON{Platform: r.Platform.String(), Info: r.Info}
			if r.Err != nil {
				out[i].Error = r.Err.Error()
			}
		}
		return output.JSON(w, out)
	}
	rows := make([][2]string, 0, len(results))
	for _, r := range results {
		if r.Err != nil {
			rows = append(rows, [2]string{r.Platform.String(), "ERROR: " + r.Err.Error()})
			continue
		}
		rows = append(rows, [2]string{r.Platform.String(), r.Info.Digest + "  " + formatBytes(r.Info.Size)})
	}
	return output.Table(w, rows)
}

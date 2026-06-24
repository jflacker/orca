// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"

	"github.com/jflacker/orca/internal/output"
	"github.com/jflacker/orca/verify"
	"github.com/spf13/cobra"
)

func verifyCmd() *cobra.Command {
	var (
		identity       string
		identityRegexp string
		issuer         string
		asJSON         bool
	)
	cmd := &cobra.Command{
		Use:   "verify <reference>",
		Short: "Verify a keyless cosign signature (Fulcio/Rekor)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if identity == "" && identityRegexp == "" {
				return fmt.Errorf("one of --certificate-identity or --certificate-identity-regexp is required")
			}
			if issuer == "" {
				return fmt.Errorf("--certificate-oidc-issuer is required when verifying a certificate identity")
			}
			ctx := cmd.Context()
			res, err := verify.Verify(ctx, args[0], verify.Identity{
				OIDCIssuer: issuer,
				SAN:        identity,
				SANRegexp:  identityRegexp,
			})
			if err != nil {
				return err
			}
			if asJSON {
				return output.JSON(cmd.OutOrStdout(), res)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "verified: %s (%s)\n", args[0], res.Digest)
			return err
		},
	}
	cmd.Flags().StringVar(&identity, "certificate-identity", "", "Exact certificate SAN to require")
	cmd.Flags().StringVar(&identityRegexp, "certificate-identity-regexp", "", "Certificate SAN regexp to require")
	cmd.Flags().StringVar(&issuer, "certificate-oidc-issuer", "", "Required OIDC issuer")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

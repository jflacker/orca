// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/jflacker/orca/oci"
)

func withTestOCIOptions(ctx context.Context, opts []oci.Option) context.Context {
	return withOCIOptions(ctx, opts)
}

// startRegistryWithImage starts an in-memory registry and pushes one image per tag.
func startRegistryWithImage(t *testing.T, repo string, tags []string) (string, []oci.Option) {
	t.Helper()
	srv := httptest.NewServer(registry.New())
	t.Cleanup(srv.Close)
	host := strings.TrimPrefix(srv.URL, "http://")
	craneOpts := []crane.Option{crane.WithTransport(srv.Client().Transport), crane.Insecure}
	for _, tag := range tags {
		img, err := random.Image(1024, 2)
		if err != nil {
			t.Fatal(err)
		}
		if err := crane.Push(img, host+"/"+repo+":"+tag, craneOpts...); err != nil {
			t.Fatal(err)
		}
	}
	opts := []oci.Option{oci.WithInsecure(), oci.WithTransport(srv.Client().Transport)}
	return host, opts
}

// startRegistryWithIndex starts an in-memory registry and pushes a multi-arch index.
// Each platform's image config is set to the matching os/arch so InspectImage
// returns accurate platform metadata.
func startRegistryWithIndex(t *testing.T, repo, tag string, platforms []string) (string, []oci.Option) {
	t.Helper()
	srv := httptest.NewServer(registry.New())
	t.Cleanup(srv.Close)
	host := strings.TrimPrefix(srv.URL, "http://")

	adds := make([]mutate.IndexAddendum, 0, len(platforms))
	for _, p := range platforms {
		parts := strings.SplitN(p, "/", 2)
		osName := parts[0]
		archName := parts[1]

		base, err := random.Image(1024, 2)
		if err != nil {
			t.Fatal(err)
		}
		cfg, err := base.ConfigFile()
		if err != nil {
			t.Fatal(err)
		}
		cfg.OS = osName
		cfg.Architecture = archName
		img, err := mutate.ConfigFile(base, cfg)
		if err != nil {
			t.Fatal(err)
		}

		adds = append(adds, mutate.IndexAddendum{
			Add: img,
			Descriptor: v1.Descriptor{
				Platform: &v1.Platform{
					OS:           osName,
					Architecture: archName,
				},
			},
		})
	}
	idx := mutate.AppendManifests(empty.Index, adds...)

	r, err := name.ParseReference(host+"/"+repo+":"+tag, name.Insecure)
	if err != nil {
		t.Fatal(err)
	}
	if err := remote.WriteIndex(r, idx, remote.WithTransport(srv.Client().Transport)); err != nil {
		t.Fatal(err)
	}

	opts := []oci.Option{oci.WithInsecure(), oci.WithTransport(srv.Client().Transport)}
	return host, opts
}

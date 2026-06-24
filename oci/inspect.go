// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Platform identifies an OS/arch/variant triple.
type Platform struct {
	OS, Architecture, Variant string
}

// String returns the platform as "os/arch" or "os/arch/variant".
func (p Platform) String() string {
	s := p.OS + "/" + p.Architecture
	if p.Variant != "" {
		s += "/" + p.Variant
	}
	return s
}

// ImageInfo is the inspected metadata of a single-platform image.
type ImageInfo struct {
	Reference    string    `json:"reference"`
	Digest       string    `json:"digest"`
	MediaType    string    `json:"media_type"`
	OS           string    `json:"os"`
	Architecture string    `json:"architecture"`
	Created      time.Time `json:"created"`
	Layers       int       `json:"layers"`
	Size         int64     `json:"size_bytes"`
}

// IndexSummary describes a multi-platform image index.
type IndexSummary struct {
	Reference string     `json:"reference"`
	Digest    string     `json:"digest"`
	MediaType string     `json:"media_type"`
	Platforms []Platform `json:"platforms"`
}

// get fetches the raw descriptor for ref without resolving an index to a single platform.
func (c *Client) get(ctx context.Context, ref string) (*remote.Descriptor, error) {
	r, err := c.parseRef(ref)
	if err != nil {
		return nil, fmt.Errorf("parsing reference %q: %w", ref, err)
	}
	desc, err := remote.Get(r, c.remoteOpts(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("fetching %q: %w", ref, err)
	}
	return desc, nil
}

// IsIndex reports whether ref points at a multi-platform image index.
func (c *Client) IsIndex(ctx context.Context, ref string) (bool, error) {
	desc, err := c.get(ctx, ref)
	if err != nil {
		return false, err
	}
	return desc.MediaType.IsIndex(), nil
}

// IndexSummary lists the platforms in an image index, skipping attestation entries
// (those with a nil platform or OS == "unknown").
func (c *Client) IndexSummary(ctx context.Context, ref string) (*IndexSummary, error) {
	desc, err := c.get(ctx, ref)
	if err != nil {
		return nil, err
	}
	idx, err := desc.ImageIndex()
	if err != nil {
		return nil, fmt.Errorf("reading index %q: %w", ref, err)
	}
	manifest, err := idx.IndexManifest()
	if err != nil {
		return nil, fmt.Errorf("reading index manifest %q: %w", ref, err)
	}
	sum := &IndexSummary{
		Reference: ref,
		Digest:    desc.Digest.String(),
		MediaType: string(desc.MediaType),
	}
	for _, m := range manifest.Manifests {
		if m.Platform == nil || m.Platform.OS == "unknown" {
			continue
		}
		sum.Platforms = append(sum.Platforms, Platform{
			OS:           m.Platform.OS,
			Architecture: m.Platform.Architecture,
			Variant:      m.Platform.Variant,
		})
	}
	return sum, nil
}

// InspectImage returns metadata for a single image. If ref is an index and
// platform is non-nil, the matching child image is inspected.
func (c *Client) InspectImage(ctx context.Context, ref string, platform *v1.Platform) (*ImageInfo, error) {
	opts := c.remoteOpts(ctx)
	if platform != nil {
		opts = append(opts, remote.WithPlatform(*platform))
	}
	r, err := c.parseRef(ref)
	if err != nil {
		return nil, fmt.Errorf("parsing reference %q: %w", ref, err)
	}
	img, err := remote.Image(r, opts...)
	if err != nil {
		return nil, fmt.Errorf("fetching image %q: %w", ref, err)
	}
	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("computing digest: %w", err)
	}
	mt, err := img.MediaType()
	if err != nil {
		return nil, fmt.Errorf("reading media type: %w", err)
	}
	cfg, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("reading layers: %w", err)
	}
	var size int64
	for i, l := range layers {
		s, err := l.Size()
		if err != nil {
			return nil, fmt.Errorf("reading size of layer %d: %w", i, err)
		}
		size += s
	}
	created := cfg.Created.Time
	return &ImageInfo{
		Reference:    ref,
		Digest:       digest.String(),
		MediaType:    string(mt),
		OS:           cfg.OS,
		Architecture: cfg.Architecture,
		Created:      created,
		Layers:       len(layers),
		Size:         size,
	}, nil
}

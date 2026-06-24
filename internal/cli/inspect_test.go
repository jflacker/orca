// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jflacker/orca/oci"
)

func TestInspectIndexShowsSummaryByDefault(t *testing.T) {
	host, opts := startRegistryWithIndex(t, "idx", "1.0", []string{"linux/amd64", "linux/arm64"})
	cmd := New()
	cmd.SetArgs([]string{"inspect", host + "/idx:1.0"})
	cmd.SetContext(withTestOCIOptions(context.Background(), opts))
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("inspect: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "image index") || !strings.Contains(s, "linux/amd64") {
		t.Fatalf("index summary missing, got: %s", s)
	}
}

func TestInspectImageWithPlatform(t *testing.T) {
	host, opts := startRegistryWithIndex(t, "idx", "1.0", []string{"linux/amd64", "linux/arm64"})
	cmd := New()
	cmd.SetArgs([]string{"inspect", host + "/idx:1.0", "--platform", "linux/arm64", "--json"})
	cmd.SetContext(withTestOCIOptions(context.Background(), opts))
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if !strings.Contains(out.String(), "\"architecture\": \"arm64\"") {
		t.Fatalf("expected arm64 image, got: %s", out.String())
	}
}

func TestInspectAllPlatformsText(t *testing.T) {
	host, opts := startRegistryWithIndex(t, "idx", "1.0", []string{"linux/amd64", "linux/arm64"})
	cmd := New()
	cmd.SetArgs([]string{"inspect", host + "/idx:1.0", "--all-platforms"})
	cmd.SetContext(withTestOCIOptions(context.Background(), opts))
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("inspect --all-platforms: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "linux/amd64") {
		t.Fatalf("expected linux/amd64 in output, got: %s", s)
	}
	if !strings.Contains(s, "linux/arm64") {
		t.Fatalf("expected linux/arm64 in output, got: %s", s)
	}
}

func TestRenderAllPlatformsJSONIncludesError(t *testing.T) {
	var b bytes.Buffer
	results := []oci.PlatformResult{
		{Platform: oci.Platform{OS: "linux", Architecture: "arm64"}, Err: errors.New("boom")},
		{Platform: oci.Platform{OS: "linux", Architecture: "amd64"}, Info: &oci.ImageInfo{Digest: "sha256:abc"}},
	}
	if err := renderAllPlatforms(&b, results, true); err != nil {
		t.Fatal(err)
	}
	got := b.String()
	if !strings.Contains(got, "boom") {
		t.Fatalf("expected error string \"boom\" in JSON output, got: %s", got)
	}
	if !strings.Contains(got, "sha256:abc") {
		t.Fatalf("expected digest \"sha256:abc\" in JSON output, got: %s", got)
	}
	if strings.Contains(got, "\"error\": null") || strings.Contains(got, ": null") {
		t.Fatalf("JSON output must not contain null for error field, got: %s", got)
	}
}

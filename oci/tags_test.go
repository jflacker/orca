// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"
	"reflect"
	"sort"
	"testing"
)

func TestFilterAndLimit(t *testing.T) {
	tags := []string{"1.0", "1.1", "2.0", "latest"}
	if got := FilterAndLimit(tags, "1.", 0); !reflect.DeepEqual(got, []string{"1.0", "1.1"}) {
		t.Fatalf("filter only: got %v", got)
	}
	if got := FilterAndLimit(tags, "", 2); !reflect.DeepEqual(got, []string{"1.0", "1.1"}) {
		t.Fatalf("limit only: got %v", got)
	}
	// Filter is applied BEFORE limit.
	if got := FilterAndLimit(tags, "1.", 1); !reflect.DeepEqual(got, []string{"1.0"}) {
		t.Fatalf("filter-then-limit: got %v", got)
	}
}

func TestListTagsAgainstRegistry(t *testing.T) {
	host, opts, _ := newTestRegistry(t)
	pushImage(t, host+"/repo:1.0", opts...)
	pushImage(t, host+"/repo:1.1", opts...)

	c := NewClient(WithCraneOptions(opts...))
	got, err := c.ListTags(context.Background(), host+"/repo")
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, []string{"1.0", "1.1"}) {
		t.Fatalf("want tags [1.0 1.1], got %v", got)
	}
}

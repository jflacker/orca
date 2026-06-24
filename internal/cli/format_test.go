// SPDX-License-Identifier: Apache-2.0

package cli

import "testing"

func TestFormatBytes(t *testing.T) {
	cases := map[int64]string{0: "0 B", 1024: "1.00 KB", 1048576: "1.00 MB"}
	for in, want := range cases {
		if got := formatBytes(in); got != want {
			t.Fatalf("formatBytes(%d) = %q, want %q", in, got, want)
		}
	}
}

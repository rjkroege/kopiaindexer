// write some tests...

package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseTheStream(t *testing.T) {

	for _, k := range []struct {
		input      string
		want       string
		snapshotid string
	}{
		{
			// Two lines.
			input: `k7d4987f893573278f5584400a47d1ac8  k76d7b2df28ab5a559e15e8aa7c319500/a/
k71691dad06d9c9f975369373bcd6e413  k76d7b2df28ab5a559e15e8aa7c319500/a/.git/
`,
			want: `k7d4987f893573278f5584400a47d1ac8 XXX k76d7b2df28ab5a559e15e8aa7c319500 /a/
k71691dad06d9c9f975369373bcd6e413 XXX k76d7b2df28ab5a559e15e8aa7c319500 /a/.git/
`,
			snapshotid: "k76d7b2df28ab5a559e15e8aa7c319500",
		},
		{
			// One line.
			input: `k7d4987f893573278f5584400a47d1ac8  k76d7b2df28ab5a559e15e8aa7c319500/a/
`,
			want: `k7d4987f893573278f5584400a47d1ac8 XXX k76d7b2df28ab5a559e15e8aa7c319500 /a/
`,
			snapshotid: "k76d7b2df28ab5a559e15e8aa7c319500",
		},
		{
			// Extra space after the hash.
			input: `k7d4987f893573278f5584400a47d1ac8         k76d7b2df28ab5a559e15e8aa7c319500/a/
`,
			want: `k7d4987f893573278f5584400a47d1ac8 XXX k76d7b2df28ab5a559e15e8aa7c319500 /a/
`,
			snapshotid: "k76d7b2df28ab5a559e15e8aa7c319500",
		},
		{
			// Spaces in file name. Note 1 space before the and 2 between the and quick.
			// Test is currently invalid. There should be 2 escaped spaces between the and quick.
			input: `k7d4987f893573278f5584400a47d1ac8 k76d7b2df28ab5a559e15e8aa7c319500/a/ the  quick
`,
			want: `k7d4987f893573278f5584400a47d1ac8 XXX k76d7b2df28ab5a559e15e8aa7c319500 /a/%20the%20quick
`,
			snapshotid: "k76d7b2df28ab5a559e15e8aa7c319500",
		},
		{
			// Newlines in file name. Note 1 space before the and 2 between the and quick.
			// Test is currently invalid. The newline is replaced with a space.
			input: `k7d4987f893573278f5584400a47d1ac8 k76d7b2df28ab5a559e15e8aa7c319500/a/ the
quick fox
`,
			want: `k7d4987f893573278f5584400a47d1ac8 XXX k76d7b2df28ab5a559e15e8aa7c319500 /a/%20the%20quick%20fox
`,
			snapshotid: "k76d7b2df28ab5a559e15e8aa7c319500",
		},
	} {

		writer := new(bytes.Buffer)
		reader := strings.NewReader(k.input)
		if err := parseTheStream(reader, k.snapshotid, "XXX", writer); err != nil {
			t.Errorf("parse shouldn't have made error: %v", err)
		}
		want := k.want
		got := string(writer.Bytes())
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("output mismatch (-want +got):\n%s", diff)
		}
	}
}

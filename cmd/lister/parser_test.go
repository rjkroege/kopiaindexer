package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
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
			want: `k7d4987f893573278f5584400a47d1ac8 XXX k76d7b2df28ab5a559e15e8aa7c319500 /a/%20the%20%20quick
`,
			snapshotid: "k76d7b2df28ab5a559e15e8aa7c319500",
		},
		{
			// Newlines in file name.
			input: `k7d4987f893573278f5584400a47d1ac8 k76d7b2df28ab5a559e15e8aa7c319500/a/ the
quick fox
`,
			want: `k7d4987f893573278f5584400a47d1ac8 XXX k76d7b2df28ab5a559e15e8aa7c319500 /a/%20the%0Aquick%20fox
`,
			snapshotid: "k76d7b2df28ab5a559e15e8aa7c319500",
		},
		{
			input: `f7baa4fbb7ab4719e75ec7c3377ecf0f   f03e4132dfe5398d579fc910cb362c9a/March_piano_songs/Beethoven - Fur Elise (original) sheet music for Piano - 8notes.com_files/x64_min.css`,
			want: `f7baa4fbb7ab4719e75ec7c3377ecf0f XXX f03e4132dfe5398d579fc910cb362c9a /March_piano_songs/Beethoven%20-%20Fur%20Elise%20%28original%29%20sheet%20music%20for%20Piano%20-%208notes.com_files/x64_min.css
`,
			snapshotid: "f03e4132dfe5398d579fc910cb362c9a",
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

func TestParseLargeStream(t *testing.T) {
	testdatafile := filepath.Join("testdata", "medium.input")
	reader, err := os.Open(testdatafile)
	if err != nil {
		t.Fatalf("can't open testdata %q: %v", testdatafile, err)
	}
	defer reader.Close()

	writername := filepath.Join("testdata", "medium.output.maybe")
	writer, err := os.Create(writername)
	defer writer.Close()

	// Actually run the test.
	if err := parseTheStream(reader, "f03e4132dfe5398d579fc910cb362c9a", "XXX", writer); err != nil {
		t.Errorf("parse shouldn't have made error: %v", err)
	}

	// Line-oriented diff of the want and got.
	wantname := filepath.Join("testdata", "medium.output")
	wantfd, err := os.Open(wantname)
	if err != nil {
		t.Errorf("missing baseline %q -- rebase please?", wantname)
		return
	}
	defer wantfd.Close()

	wantscanner := bufio.NewScanner(wantfd)
	want := make([]string, 0)
	for wantscanner.Scan() {
		want = append(want, wantscanner.Text())
	}
	if err := wantscanner.Err(); err != nil {
		t.Errorf("can't read want files %q: %v", wantname, err)
	}

	if _, err := writer.Seek(0, 0); err != nil {
		t.Errorf("can't go back to start of output %q: %v", writername, err)
	}

	gotscanner := bufio.NewScanner(writer)
	got := make([]string, 0)
	for gotscanner.Scan() {
		got = append(got, gotscanner.Text())
	}
	if err := gotscanner.Err(); err != nil {
		t.Errorf("can't read got files %q: %v", writername, err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("output mismatch (-want +got):\n%s", diff)
	} else {
		// They're the same. Don't need the results.
		os.Remove(writername)
	}
}

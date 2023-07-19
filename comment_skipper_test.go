package main

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

func TestCommentSkipper(t *testing.T) {
	for _, k := range []struct {
		input string
		want  string
	}{

		{
			input: "{",
			want:  "{",
		},
		{
			input: "// foo\n{",
			want:  "{",
		},
	} {
		buffy := bufio.NewReader(strings.NewReader(k.input))

		SkipCommentLines(buffy)
		rest, err := io.ReadAll(buffy)

		if err != nil {
			t.Errorf("reading the rest didn't work: %v", err)
		}

		if got, want := string(rest), k.want; got != want {
			t.Errorf("fooey! got %v, want %v", got, want)
		}

	}

}

package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLexer(t *testing.T) {
	for _, k := range []struct {
		input string
		want  []Token
	}{
		{
			"",
			[]Token{},
		},
		{
			" ",
			[]Token{
				{
					Kind:     tWhitespace,
					Contents: " ",
				},
			},
		},
		{
			"aaa",
			[]Token{
				{
					Kind:     tText,
					Contents: "aaa",
				},
			},
		},
		{
			"ab x",
			[]Token{
				{
					Kind:     tText,
					Contents: "ab",
				},
				{
					Kind:     tWhitespace,
					Contents: " ",
				},
				{
					Kind:     tText,
					Contents: "x",
				},
			},
		},
		{
			"ab  x ",
			[]Token{
				{
					Kind:     tText,
					Contents: "ab",
				},
				{
					Kind:     tWhitespace,
					Contents: "  ",
				},
				{
					Kind:     tText,
					Contents: "x",
				},
				{
					Kind:     tWhitespace,
					Contents: " ",
				},
			},
		},
		{
			"ab\nx",
			[]Token{
				{
					Kind:     tText,
					Contents: "ab",
				},
				{
					Kind:     tNewline,
					Contents: "\n",
				},
				{
					Kind:     tText,
					Contents: "x",
				},
			},
		},
		{
			"a aa bbb\n\n\ndddd\n eeeee\n       ferns \ngrasses   long\n",
			[]Token{
				{Kind: 2, Contents: "a"},
				{Kind: 1, Contents: " "},
				{Kind: 2, Contents: "aa"},
				{Kind: 1, Contents: " "},
				{Kind: 2, Contents: "bbb"},
				{Kind: 4, Contents: "\n"},
				{Kind: 4, Contents: "\n"},
				{Kind: 4, Contents: "\n"},
				{Kind: 2, Contents: "dddd"},
				{Kind: 4, Contents: "\n"},
				{Kind: 1, Contents: " "},
				{Kind: 2, Contents: "eeeee"},
				{Kind: 4, Contents: "\n"},
				{Kind: 1, Contents: "       "},
				{Kind: 2, Contents: "ferns"},
				{Kind: 1, Contents: " "},
				{Kind: 4, Contents: "\n"},
				{Kind: 2, Contents: "grasses"},
				{Kind: 1, Contents: "   "},
				{Kind: 2, Contents: "long"},
				{Kind: 4, Contents: "\n"},
			},
		},
	} {
		t.Logf("lexing %q", k.input)
		lexer := NewLexer(strings.NewReader(k.input))
		got := make([]Token, 0)
		for lexer.Scan() {
			tok := lexer.NextToken()
			t.Logf("%d %q", tok.Kind, tok.Contents)
			got = append(got, tok)
		}

		want := k.want
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("output mismatch (-want +got):\n%s", diff)
		}

	}

}

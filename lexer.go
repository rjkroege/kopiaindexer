package main

import (
	"bufio"
	"io"
	"strings"
)

const (
	tEmpty      = 0
	tWhitespace = iota
	tText
	tHash
	tNewline
)

const (
	tEof   = -1
	tError = -2
)

type Token struct {
	Kind     int
	Contents string
}

type Lexer struct {
	r       *bufio.Reader
	err     error
	Kind    int
	b       strings.Builder
	pending Token
}

// API and structure is cribbed from Go buifio.Scanner implementation.
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		r: bufio.NewReader(r),
	}
}

func (s *Lexer) Scan() bool {
	// Scanner loop

	if s.Kind == tEof || s.Kind == tError {
		return false
	}

	for {
		b, err := s.r.ReadByte()
		if err != nil {
			if s.Kind != tEmpty && s.Kind != tNewline {
				s.pending = Token{
					Kind:     s.Kind,
					Contents: s.b.String(),
				}
				s.b.Reset()
				s.Kind = tError
				if err == io.EOF {
					s.Kind = tEof
				}

				return true
			}
			s.Kind = tError
			if err == io.EOF {
				s.Kind = tEof
			}
			return false
		}

		switch b {
		case ' ':
			if s.Kind == tText {
				s.pending = Token{
					Kind:     s.Kind,
					Contents: s.b.String(),
				}
				s.b.Reset()
				s.Kind = tWhitespace
				s.r.UnreadByte()
				return true
			}
			s.Kind = tWhitespace
			s.b.WriteByte(b)
		case '\n':
			if s.Kind == tWhitespace || s.Kind == tText {
				s.pending = Token{
					Kind:     s.Kind,
					Contents: s.b.String(),
				}
				s.b.Reset()
				s.Kind = tNewline
				s.r.UnreadByte()
				return true
			}
			s.Kind = tNewline
			s.pending = Token{
				Kind:     s.Kind,
				Contents: "\n",
			}
			return true
		default:
			if s.Kind == tWhitespace {
				s.pending = Token{
					Kind:     s.Kind,
					Contents: s.b.String(),
				}
				s.b.Reset()
				s.Kind = tText
				s.r.UnreadByte()
				return true
			}
			s.b.WriteByte(b)
			s.Kind = tText
		}
	}

	return false
}

func (s *Lexer) NextToken() Token {
	if s.Kind == tEmpty {
		panic("no token when calling NextToken")
	}

	return s.pending
}

func (s *Lexer) Err() error {
	if s.Kind == tError {
		return s.err
	}
	return nil
}

// write some tests...

package main

import (
	"testing"
	"os"
	"strings"
)


const input = `k7d4987f893573278f5584400a47d1ac8  k76d7b2df28ab5a559e15e8aa7c319500/a/
k71691dad06d9c9f975369373bcd6e413  k76d7b2df28ab5a559e15e8aa7c319500/a/.git/
`

func TestParseTheStream(t *testing.T) {

	// TODO(rjk): Pass in a writer.

	reader := strings.NewReader(input)
	if err := parseTheStream(reader, "k76d7b2df28ab5a559e15e8aa7c319500", os.Stdout); err != nil {
		t.Errorf("parse shouldn't have made error: %v", err)
	}

}
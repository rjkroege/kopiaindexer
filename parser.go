package main

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/edwingeng/deque/v2"
)

func pathUrlEscape(filename string) string {
	pathparts := strings.Split(filename, string(os.PathSeparator))
	escapedpathparts := make([]string, 0, len(pathparts))
	for _, s := range pathparts {
		escapedpathparts = append(escapedpathparts, url.PathEscape(s))
	}
	finalpath := strings.Join(escapedpathparts, string(os.PathSeparator))
	return finalpath
}

func printEntry(dq *deque.Deque[Token], writer io.Writer, snapshotid, escapedsource string) error {
	hashcode := dq.PopFront()
	if hashcode.Kind != tHash {
		return fmt.Errorf("parse error?")
	}

	sb := new(strings.Builder)
	for _, t := range dq.DequeueMany(-1) {
		sb.WriteString(t.Contents)
	}
	finalpath := pathUrlEscape(sb.String())

	// Assembly a final entry.
	buffy := new(bytes.Buffer)
	buffy.WriteString(hashcode.Contents)
	buffy.WriteByte(' ')
	buffy.WriteString(escapedsource)
	buffy.WriteByte(' ')
	buffy.WriteString(snapshotid)
	buffy.WriteByte(' ')
	buffy.WriteString(finalpath)
	buffy.WriteByte('\n')

	// The call to Write is atomic.
	if _, err := writer.Write(buffy.Bytes()); err != nil {
		return err
	}
	return nil
}

// parseTheStream extracts the components (hash, filename) from the
// output of a `kopia ls -o`. Because files names can contain whitespace
// (including newlines sigh), it uses a push-down automoton approach to
// detect the hash and filename components. This approach will fail if
// the snapshotid is whitespace-prefixed in the filename itself.
func parseTheStream(cmdout io.Reader, snapshotid, escapedsource string, writer io.Writer) error {
	lexer := NewLexer(cmdout)
	dq := deque.NewDeque[Token]()

	for lexer.Scan() {
		token := lexer.NextToken()

		if strings.HasPrefix(token.Contents, snapshotid) {
			// If prefix of word is snapshotid then a valid parse would be '\n'? hash space
			s := dq.PopBack()
			h := dq.PopBack()
			dq.TryPopBack()

			if s.Kind != tWhitespace || h.Kind != tText {
				// TODO(rjk): Some kind of details about the parse error.
				return fmt.Errorf("parse error")
			}

			memoizedentry := Token{
				Kind:     tHash,
				Contents: h.Contents,
			}

			// If we are on the first line of output, we (as of yet) have no idea
			// where the file name will end. So just skip the pop.
			if dq.Len() > 0 {
				if err := printEntry(dq, writer, snapshotid, escapedsource); err != nil {
					return err
				}
			}

			// Push the memo
			dq.PushBack(memoizedentry)

			// Push the filename
			dq.PushBack(Token{
				Kind:     tText,
				Contents: strings.TrimPrefix(token.Contents, snapshotid),
			})
		} else {
			dq.PushBack(token)
		}
	}
	if err := lexer.Err(); err != nil {
		return fmt.Errorf("reading input: %v", err)
	}

	// I have an unhandled example.
	if dq.Len() > 0 {
		// I assume that Kopia is going to always produce files where each line ends
		// with a newline token. Remove that.
		nl := dq.PopBack()
		if nl.Kind != tNewline {
			dq.PushBack(nl)
		}

		if err := printEntry(dq, writer, snapshotid, escapedsource); err != nil {
			return nil
		}
	}
	return nil
}

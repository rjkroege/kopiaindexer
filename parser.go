package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/edwingeng/deque/v2"
)

func pathUrlEscape(filename []byte) string {
	pathparts := bytes.Split(filename, []byte{os.PathSeparator})
	escapedpathparts := make([]string, 0, len(pathparts))
	for _, s := range pathparts {
		escapedpathparts = append(escapedpathparts, url.PathEscape(string(s)))
	}
	finalpath := strings.Join(escapedpathparts, string(os.PathSeparator))
	return finalpath
}

func printEntry(dq *deque.Deque[[]byte], writer io.Writer, snapshotid, escapedsource string) error {
	hashcode := dq.PopFront()

	// TODO(rjk): Converts all in-filename whitespace into a single space.
	filename := bytes.Join(dq.DequeueMany(-1), []byte(" "))
	finalpath := pathUrlEscape(filename)

	// Assembly a final entry.
	buffy := new(bytes.Buffer)
	buffy.Write(hashcode)
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
	scanner := bufio.NewScanner(cmdout)
	// Set the split function for the scanning operation.
	scanner.Split(bufio.ScanWords)

	dq := deque.NewDeque[[]byte]()

	for scanner.Scan() {
		t := scanner.Bytes()
		token := make([]byte, len(t))
		copy(token, t)

		// TODO(rjk): Preserve the whitespace in a robust way.
		if bytes.HasPrefix(token, []byte(snapshotid)) {
			// If prefix of word is snapshotid then head(stack) is the current entry.
			// So memoize it
			memoizedentry := dq.PopBack()

			// If we are on the first line of output, we (as of yet) have no idea
			// where the file name will end. So just skip the pop.
			if dq.Len() > 0 {
				if err := printEntry(dq, writer, snapshotid, escapedsource); err != nil {
					return nil
				}
			}

			// Push the memo
			dq.PushBack(memoizedentry)

			// Push the filename
			dq.PushBack(bytes.TrimPrefix(token, []byte(snapshotid)))
		} else {
			dq.PushBack(token)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading input:", err)
	}

	// I have an unhandled example.
	if dq.Len() > 0 {
		if err := printEntry(dq, writer, snapshotid, escapedsource); err != nil {
			return nil
		}
	}
	return nil
}

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/edwingeng/deque/v2"
)

func printEntry(dq *deque.Deque[[]byte], writer io.Writer) error {
	hashcode := dq.PopFront()
	filenameparts := dq.DequeueMany(-1)

	allpieces := make([][]byte, 0, len(filenameparts)+1)
	allpieces = append(allpieces, hashcode)
	allpieces = append(allpieces, filenameparts...)

	// TODO(rjk): Fix for the final output format.
	totallen := 0
	for _, fnp := range allpieces {
		totallen += len(fnp)
		totallen++
	}
	totallen++
	// TODO(rjk): I have been too clever here. I will need to fix this up for
	// the URL-style encoding.

	entry := make([]byte, 0, totallen)
	for _, fnp := range allpieces {
		entry = append(entry, fnp...)
		entry = append(entry, ' ')
	}
	entry = append(entry, '\n')

	// The call to Write is atomic.
	if _, err := writer.Write(entry); err != nil {
		return err
	}
	return nil
}

// parseTheStream extracts the components (hash, filename) from the
// output of a `kopia ls -o`. Because files names can contain whitespace
// (including newlines sigh), it uses a push-down automoton approach to
// detect the hash and filename components. This approach will fail if
// the snapshotid is whitespace-prefixed in the filename itself.
func parseTheStream(cmdout io.Reader, snapshotid string, writer io.Writer) error {
	scanner := bufio.NewScanner(cmdout)
	// Set the split function for the scanning operation.
	scanner.Split(bufio.ScanWords)

	dq := deque.NewDeque[[]byte]()

	for scanner.Scan() {
		token := scanner.Bytes()
		// DEBUG. Remove.
		fmt.Println("token:", string(token))
		// TODO(rjk): Preserve the whitespace in a robust way.

		if bytes.HasPrefix(token, []byte(snapshotid)) {
			// If prefix of word is snapshotid then head(stack) is the current entry.
			// So memoize it
			memoizedentry := dq.PopBack()

			// If we are on the first line of output, we (as of yet) have no idea
			// where the file name will end. So just skip the pop.
			if dq.Len() > 0 {
				if err := printEntry(dq, writer); err != nil {
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
		if err := printEntry(dq, writer); err != nil {
			return nil
		}
	}

	return nil
}

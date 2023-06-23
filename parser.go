package main

import (
	"bufio"
	"io"
	"fmt"
	"os"
	"bytes"

	"github.com/edwingeng/deque/v2"
)

const (
	maybe_entry_start = iota
	maybe_snapshot_id
	hash_snapshot_separation
	filename
)

// need a token of lookback. 
// tokens are separated by whitespace.
// push tokens into stack
// if current token starts with snapshotid then
//	previous token is the key
// need to read byte-by-byte
// if I find the snapshotid, then the most recent space-free token 
// states:
// before (accumulating bytes in the key)
// found a 


func printEntry(dq *deque.Deque[[]byte], writer io.Writer) error {
	hashcode := dq.PopFront()
	filenameparts := dq.DequeueMany(-1)
	
	allpieces := make([][]byte, 0, len(filenameparts) + 1)
	allpieces = append(allpieces, hashcode)
	allpieces = append(allpieces, filenameparts...)

	// TODO(rjk): Fix for the final output format.
	totallen := 0
	for _, fnp := range allpieces {
		totallen += len(fnp)
		totallen ++
	}
	totallen ++
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

// this is a classic PDA
// but I need to record the separations.
// to handle "the  quick" (with two spaces)

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


/*
get-word (with its whitepsace)
if prefix of word is snapshotid
	then head(stack) is the current entry
pop last stack word (this is a filehash)
pop all pushed words:
	deepest is the entry
	print the remaining as the filename (escaped)
push the memoized stack word (filehash)
*/
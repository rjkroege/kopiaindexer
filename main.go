package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/kopia/kopia/cli"
)

// Expects a single manifest file as argument.

func main() {
	log.Println("foo")

	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatalf("expect a single manifest file as argument\n")
	}

	mfn := flag.Args()[0]

	fd, err := os.Open(mfn)
	if err != nil {
		log.Fatalf("given argument %q can't be opened: %v", mfn, err)
	}

	bfd := bufio.NewReader(fd)

	// Manifest files start with a comment block. This is not valid JSON. So
	// before we decode, we need to advance the fd to the start of the first
	// line following the comment block.

	SkipCommentLines(bfd)

	// Decode what remains.
	dec := json.NewDecoder(bfd)
	dec.DisallowUnknownFields()

	// The manifest is a snapshot manifest
	var k cli.SnapshotManifest
	if err := dec.Decode(&k); err != nil {
		log.Fatal(err)
	}

	log.Println(k)

	escapedsource := pathUrlEscape(k.Source.String())
	log.Println("List snapshot", k.RootEntry.ObjectID)
	if k.RootEntry.DirSummary.TotalFileCount == 1 {
		id := k.RootEntry.ObjectID.String()
		fmt.Printf("%s %s %s %s\n", id, escapedsource, id, escapedsource)
		return
	}
	// TODO(rjk): Support single files manifests.
	listSnapshot(k.RootEntry.ObjectID.String(), escapedsource)
	return
}

func listSnapshot(snapshotid, escapedsource string) {
	cmd := exec.Command("kopia", "list", "-r", "-o", snapshotid)
	cmdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	if err := parseTheStream(cmdout, snapshotid, escapedsource, os.Stdout); err != nil {
		log.Fatal("parseTheStream failed:", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

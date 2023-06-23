package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/kopia/kopia/cli"
)

func main() {
	log.Println("foo")

	// Get the list of snapshots.
	cmd := exec.Command("kopia", "snapshot", "list", "-a", "--json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	dec := json.NewDecoder(stdout)
	dec.DisallowUnknownFields()

	var manifests []cli.SnapshotManifest

	if err := dec.Decode(&manifests); err != nil {
		log.Fatal(err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	for _, k := range manifests {
		log.Println(k.ID, k.Source)

		// these can run in parallel
		// is there a benefit?
		listSnapshot(string(k.ID))
	}
}

func listSnapshot(snapshotid string) {
	// Get the list.

	cmd := exec.Command("kopia", "list", "-r", "-o", snapshotid)
	cmdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Copy from the stdout...
	// TODO(rjk): parse the listing.
	io.Copy(os.Stdout, cmdout)

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}


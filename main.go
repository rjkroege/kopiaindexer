package main

import (
	"encoding/json"
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
		log.Println(k.ID, k.Source.String())

		// TODO(rjk): This could be parallelized.
		escapedsource := pathUrlEscape(k.Source.String())
		listSnapshot(string(k.ID), escapedsource)
	}
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

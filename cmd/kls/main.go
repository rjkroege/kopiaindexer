package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/kopia/kopia/fs/localfs"
	"github.com/kopia/kopia/snapshot"
)

func GetDirEntry(path string) (*snapshot.DirEntry, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	php := path
	if fi.IsDir() {
		php = filepath.Join(path, localfs.ShallowEntrySuffix)
	}
	// log.Println("php", php)
	return dirEntryFromPlaceholder(php)
}

// Copied out of kopia/fs/localfs/shallow_fs.go
func dirEntryFromPlaceholder(path string) (*snapshot.DirEntry, error) {
	// TODO(rjk): Probably broken on windows.
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "dirEntryFromPlaceholder reading placeholder")
	}

	direntry := &snapshot.DirEntry{}
	buffy := bytes.NewBuffer(b)
	decoder := json.NewDecoder(buffy)

	if err := decoder.Decode(direntry); err != nil {
		return nil, errors.Wrap(err, "dirEntryFromPlaceholder JSON decoding")
	}
	return direntry, nil
}

var lFlag = flag.Bool("l", false, "long format")
var rFlag = flag.Bool("r", false, "recursive listing")

func main() {
	flag.Parse()

	cmd := []string{
		"ls",
	}

	if *lFlag {
		cmd = append(cmd, "-l")
	}

	if *rFlag {
		cmd = append(cmd, "-r")
	}

	for _, sf := range flag.Args() {
		// log.Println("sf", sf)
		de, err := GetDirEntry(sf)
		if err != nil {
			log.Println("couldn't open the placeholder", err)
			continue
		}
		io.WriteString(os.Stdout, sf+"\n")

		rcmd := append(cmd, de.ObjectID.String())
		kopia := exec.Command("/usr/local/bin/kopia", rcmd...)
		spew, err := kopia.CombinedOutput()
		if err != nil {
			log.Println("kls can't run kopia", err, "spew:", string(spew))
			continue
		}
		if _, err := os.Stdout.Write(spew); err != nil {
			// If I can't write to stdout, it's probalby a fatal
			log.Fatal("kls can't output", err)
		}
	}
}

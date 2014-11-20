package main

import (
	"fmt"
	"os"

	"github.com/gkristic/tech-meetup-2014/digest"
	"github.com/gkristic/tech-meetup-2014/path"
)

const maxFiles = 10

func main() {
	root := "."

	if len(os.Args) == 2 {
		root = os.Args[1]
	} else if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [<root>]\n", os.Args[0])
		os.Exit(1)
	}

	w := path.NewWalkerC(digestFile, digest.Dir, maxFiles)

	if _, err := w.Walk(root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	showDuplicates()
}

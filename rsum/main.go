package main

import (
	"fmt"
	"os"

	"github.com/gkristic/tech-meetup-2014/digest"
	"github.com/gkristic/tech-meetup-2014/path"
)

func main() {
	files := []string{"."}

	if len(os.Args) > 1 {
		files = os.Args[1:]
	}

	w := path.NewWalker(digest.File, digest.Dir)

	for _, fn := range files {
		if s, err := w.Walk(fn); err == nil {
			fmt.Printf("%s  %s\n", s, fn)
		} else {
			fmt.Fprint(os.Stderr, fn, ": ", err, "\n")
		}
	}
}

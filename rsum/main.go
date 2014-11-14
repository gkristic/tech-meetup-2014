package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/gkristic/tech-meetup-2014/digest"
	"github.com/gkristic/tech-meetup-2014/path"
)

const maxFiles = 20

func main() {
	files := []string{"."}

	if len(os.Args) > 1 {
		files = os.Args[1:]
	}

	w := path.NewWalkerC(digest.File, digest.Dir, maxFiles)
	var wg sync.WaitGroup

	// We use this to preserve the order in which these arguments were given.
	// There's no need to synchronize access to the slice, because all routines
	// will be writing in their own position.
	type result struct {
		name string
		hash path.Result
		err  error
	}
	results := make([]result, len(files))

	for i, fn := range files {
		wg.Add(1) // Adds "one more thing" to wait for.
		results[i].name = fn

		go func(pos int, fn string) {
			defer wg.Done()

			if s, err := w.Walk(fn); err == nil {
				results[pos].hash = s
			} else {
				results[pos].err = err
			}
		}(i, fn)
	}

	// Wait for all routines to finish.
	wg.Wait()

	for _, r := range results {
		if r.err == nil {
			fmt.Printf("%s  %s\n", r.hash, r.name)
		} else {
			fmt.Fprint(os.Stderr, r.name, ": ", r.err, "\n")
		}
	}
}

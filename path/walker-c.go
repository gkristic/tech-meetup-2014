package path

import (
	"errors"
	"os"
	"path"
)

type walkerC struct {
	digestFile FileDigestor
	digestDir  DirDigestor
	open       func(fn string, quitCh <-chan struct{}) (*tokenFile, error)
}

// ErrAbort is returned from the walker's open() function when quitCh is closed.
var ErrAbort = errors.New("aborted")

// NewWalkerC returns a concurrent walker that traverses the file tree, running
// the provided sum function for every file it finds, and summing up directory
// contents with the reduce function.
func NewWalkerC(fd FileDigestor, dd DirDigestor, maxFiles uint) Walker {
	tokenCh := make(chan struct{}, maxFiles)
	for i := uint(0); i < maxFiles; i++ {
		tokenCh <- struct{}{}
	}

	return &walkerC{
		digestFile: fd,
		digestDir:  dd,
		open: func(fn string, quitCh <-chan struct{}) (*tokenFile, error) {
			// We take advantage of closures in Go to reference the tokens
			// channel (that is a variable local to NewWalkerC()).
			select {
			case <-tokenCh:
			case <-quitCh:
				return nil, ErrAbort
			}

			f, err := os.Open(fn)
			if err != nil {
				tokenCh <- struct{}{}
				return nil, err
			}
			return &tokenFile{
				File: f,
				release: func() {
					tokenCh <- struct{}{}
				},
			}, nil
		},
	}
}

func (w *walkerC) Walk(root string) (Result, error) {
	// We use the channel as a quit signal, in case one of the concurrently
	// running routines finds an error, so others can abort as soon as possible.
	// This will have no effect if the call succeeded. Note that closing
	// channels in Go is not required; "closing" is only a broadcast into the
	// channel that unblocks all readers.
	quitCh := make(chan struct{})
	defer close(quitCh)
	return w.doWalk(root, quitCh)
}

func (w *walkerC) doWalk(root string, quitCh <-chan struct{}) (Result, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return w.digestFile(root, info, func() (File, error) {
			return w.open(root, quitCh)
		})
	}

	f, err := w.open(root, quitCh)
	if err != nil {
		return nil, err
	}

	files, err := f.Readdirnames(0)
	if err != nil {
		f.Close()
		return nil, err
	}

	f.Close()

	type walkResult struct {
		result FileResult
		err    error
	}
	ch := make(chan walkResult)

	for _, item := range files {
		go func(item string) {
			result, err := w.doWalk(path.Join(root, item), quitCh)

			// No need to ask for err != nil; we can build the whole walkResult
			// in any case, cause we'll never consider r.result if r.err != nil.
			r := walkResult{
				result: FileResult{
					Name: item,
					Hash: result,
				},
				err: err,
			}

			select {
			case ch <- r:
			case <-quitCh:
			}
		}(item)
	}

	fileResults := make([]FileResult, 0, len(files))

	for i := 0; i < len(files); i++ {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		fileResults = append(fileResults, r.result)
	}

	return w.digestDir(root, info, fileResults)
}

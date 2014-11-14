package digest

import (
	"crypto/sha1"
	"errors"
	"io"
	"os"
	"sort"

	"github.com/gkristic/tech-meetup-2014/path"
)

// ErrIncompleteCopy will be triggered if, when reading a file, its contents
// was found to be different in length than that declared in the info structure
// (from the OS's stat call).
var ErrIncompleteCopy = errors.New("incomplete copy detected")

// File returns the SHA1 digest for a given file. If non-regular, it returns nil
// instead. (The file name will still contribute for the parent directory's
// digest, though, so it will get noticed in the final value.)
func File(name string, info os.FileInfo, open func() (path.File, error)) (path.Result, error) {
	if !info.Mode().IsRegular() {
		return nil, nil
	}

	f, err := open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	h := sha1.New()

	if n, err := io.Copy(h, f); err != nil {
		return nil, err
	} else if n != info.Size() {
		return nil, ErrIncompleteCopy
	}

	return Result(h.Sum(nil)), nil
}

// Dir returns the digest for a directory.
func Dir(_ string, _ os.FileInfo, fileResults []path.FileResult) (path.Result, error) {
	sort.Sort(byName(fileResults))
	h := sha1.New()

	for _, item := range fileResults {
		h.Write(append([]byte(item.Name), byte(0)))

		if item.Hash != nil {
			h.Write([]byte(item.Hash.(Result)))
		}
		h.Write([]byte{0})
	}

	return Result(h.Sum(nil)), nil
}

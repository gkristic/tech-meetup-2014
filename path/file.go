package path

import (
	"io"
	"os"
)

// A File represents an OS file, restricting operations to reading and seeking.
// This should be easier to do, but unfortunately Go's os package exports a
// concrete type, instead of an interface for File.
type File interface {
	io.ReadSeeker
	io.Closer
}

type tokenFile struct {
	*os.File
	release func()
}

func (tf *tokenFile) Close() error {
	defer tf.release()
	tf.File.Close()
	return nil
}

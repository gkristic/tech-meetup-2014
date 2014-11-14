package path

import (
	"os"
)

// Result is a generic result from digesting.
type Result interface {
	String() string
}

type (
	// A FileDigestor returns a digest for a given file. The name provided as
	// the first argument is relative to the current directory. (Notice that the
	// name at the FileInfo type is relative to the file's directory instead.)
	// The third parameter is the function to open the file. This one should be
	// used instead of os.Open(), so that appropriate limit on the number of
	// simultaneously open files is enforced.
	FileDigestor func(string, os.FileInfo, func() (File, error)) (Result, error)

	// A DirDigestor collapses results for all files in a directory, into a
	// checksum for the directory itself. It receives the directory name and
	// attributes, and a list of files within, with their respective results due
	// to the FileDigestor function. The file results slice can be modified in
	// place, it will no longer be used by the Walker after this call.
	DirDigestor func(string, os.FileInfo, []FileResult) (Result, error)
)

// A Walker traverses the file tree.
type Walker interface {
	Walk(string) (Result, error)
}

// FileResult stores the name/hash for a file.
type FileResult struct {
	Name string
	Hash Result
}

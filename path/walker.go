package path

import (
	"os"
	"path"
)

type walker struct {
	digestFile FileDigestor
	digestDir  DirDigestor
}

// NewWalker returns a walker that traverses the file tree, running the
// provided sum function for every file it finds, and summing up directory
// contents with the reduce function.
func NewWalker(fd FileDigestor, dd DirDigestor) Walker {
	return &walker{
		digestFile: fd,
		digestDir:  dd,
	}
}

func (w *walker) Walk(root string) (Result, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return w.digestFile(root, info)
	}

	f, err := os.Open(root)
	if err != nil {
		return nil, err
	}

	files, err := f.Readdirnames(0)
	if err != nil {
		f.Close()
		return nil, err
	}

	f.Close()
	fileResults := make([]FileResult, len(files))

	for i, item := range files {
		result, err := w.Walk(path.Join(root, item))
		if err != nil {
			return nil, err
		}

		fileResults[i] = FileResult{
			Name: item,
			Hash: result,
		}
	}

	return w.digestDir(root, info, fileResults)
}

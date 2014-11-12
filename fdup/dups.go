package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/gkristic/tech-meetup-2014/digest"
	"github.com/gkristic/tech-meetup-2014/path"
)

type fileNode struct {
	hash      string
	unitSize  int64
	totalSize int64
	names     []string
}

var files = map[string]*fileNode{}

func digestFile(fn string, info os.FileInfo) (path.Result, error) {
	s, err := digest.File(fn, info)
	if err != nil {
		return nil, err
	}

	key := s.String()

	if node := files[key]; node != nil {
		if node.unitSize != info.Size() {
			fmt.Fprintf(os.Stderr, "size mismatch for %s (%d vs %d), hash %s\n",
				fn, node.unitSize, info.Size(), key)
		}
		node.totalSize += info.Size()
		node.names = append(node.names, fn)
	} else {
		files[key] = &fileNode{
			hash:      key,
			unitSize:  info.Size(),
			totalSize: info.Size(),
			names:     []string{fn},
		}
	}

	return s, err
}

func showDuplicates() {
	nodes := make([]fileNode, 0, len(files))
	for _, nodePtr := range files {
		if len(nodePtr.names) > 1 {
			nodes = append(nodes, *nodePtr)
		}
	}

	sort.Sort(byTotalSize(nodes))

	for _, node := range nodes {
		fmt.Printf("Replicated contents (totals %s) at:\n", size(node.totalSize))
		sort.Sort(sort.StringSlice(node.names))
		for _, fn := range node.names {
			fmt.Print("  ", fn, "\n")
		}
	}
}

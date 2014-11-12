package main

type byTotalSize []fileNode

func (v byTotalSize) Len() int {
	return len(v)
}

func (v byTotalSize) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v byTotalSize) Less(i, j int) bool {
	return v[i].totalSize > v[j].totalSize
}

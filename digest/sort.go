package digest

import (
	"github.com/gkristic/tech-meetup-2014/path"
)

type byName []path.FileResult

func (v byName) Len() int {
	return len(v)
}

func (v byName) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v byName) Less(i, j int) bool {
	return v[i].Name < v[j].Name
}

package fileio

import (
	"os"
	"utils"
)

// Root encapsulates access to dedicated subfolder
type Root struct {
	root utils.PathGen
}

// NewRoot constructor
func NewRoot(root utils.PathGen) Root {
	return Root{
		root: root,
	}
}

// OpenFile opens file in a given root with requsted attributes
func (r Root) OpenFile(path string, flags int, mode os.FileMode) (*os.File, error) {
	return os.OpenFile(r.root.Join(path), flags, mode)
}

// Path generates absolute path for the given file in a root foolder
func (r Root) Path(name string) string {
	return r.root.Join(name)
}

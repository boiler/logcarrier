package paths

import (
	"path/filepath"
	"time"
)

// Files ...
type Files struct {
	root           string
	rotationFormat string
}

// NewFiles constructor
func NewFiles(root string, rotationFormat string) *Files {
	return &Files{
		root: root,
	}
}

// Name implementation
func (f *Files) Name(dir string, name string, group string, t time.Time) string {
	return filepath.Clean(filepath.Join(f.root, dir, name))
}

// Rotation implementation
func (f *Files) Rotation(dir string, name string, group string, t time.Time) string {
	return frmt(f.rotationFormat, f.root, dir, name, group, t)
}

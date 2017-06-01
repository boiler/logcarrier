package paths

import (
	"time"
)

// Files ...
type Files struct {
	root           string
	nameFormat     string
	rotationFormat string
}

// NewFiles constructor
func NewFiles(root, nameFormat, rotationFormat string) *Files {
	return &Files{
		root:           root,
		nameFormat:     nameFormat,
		rotationFormat: rotationFormat,
	}
}

// Name implementation
func (f *Files) Name(dir string, name string, group string, t time.Time) string {
	return frmt(f.nameFormat, f.root, dir, name, group, t)
}

// Rotation implementation
func (f *Files) Rotation(dir string, name string, group string, t time.Time) string {
	return frmt(f.rotationFormat, f.root, dir, name, group, t)
}

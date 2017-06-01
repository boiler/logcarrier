package paths

import (
	"time"
)

// Links ...
type Links struct {
	root           string
	nameFormat     string
	rotationFormat string
}

// NewLinks counstructor
func NewLinks(root, nameFormat, rotationFormat string) *Links {
	return &Links{
		root:           root,
		nameFormat:     nameFormat,
		rotationFormat: rotationFormat,
	}
}

func (l *Links) format(fr, dir, name, group string, t time.Time) string {
	return frmt(fr, l.root, dir, name, group, t)
}

// Name implementation
func (l *Links) Name(dir, name, group string, t time.Time) string {
	return l.format(l.nameFormat, dir, name, group, t)
}

// Rotation implementation
func (l *Links) Rotation(dir, rotation, group string, t time.Time) string {
	return l.format(l.rotationFormat, dir, rotation, group, t)
}

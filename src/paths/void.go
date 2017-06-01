package paths

import "time"

// Void returns empty everything
type Void bool

// Name implementation
func (v Void) Name(dir string, name string, group string, t time.Time) string {
	return ""
}

// Rotation implementation
func (v Void) Rotation(dir string, name string, group string, t time.Time) string {
	return ""
}

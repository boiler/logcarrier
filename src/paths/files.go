package paths

import (
	format "formatter"
	"path/filepath"
	"time"
)

// Files ...
type Files struct {
	root string
	name string
	link string
}

// NewFiles constructor
func NewFiles(root, name, link string) *Files {
	return &Files{
		root: root,
		name: name,
		link: link,
	}
}

func (f *Files) format(frmt, root, dir, name, group string, t time.Time) string {
	bctx := format.NewContextBuilder()
	bctx.AddString("dir", dir)
	bctx.AddString("name", name)
	bctx.AddString("group", group)
	bctx.AddTime("time", t)
	ctx, err := bctx.BuildContext()
	if err != nil {
		panic(err)
	}
	res, err := format.Format(frmt, ctx)
	if err != nil {
		panic(err)
	}
	return filepath.Clean(filepath.Join(root, res))
}

// Name implementation
func (f *Files) Name(dir string, name string, group string, t time.Time) string {
	return f.format(f.name, f.root, dir, name, group, t)
}

// Link implementation
func (f *Files) Link(dir string, name string, group string, t time.Time) string {
	return f.format(f.link, f.root, dir, name, group, t)
}

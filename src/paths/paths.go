package paths

import (
	format "formatter"
	"path/filepath"
	"time"
)

// Paths is an interface for path name generation
// and file rotation
type Paths interface {
	Name(dir, name, group string, t time.Time) string
	Rotation(dir, name, group string, t time.Time) string
}

// frmt generic formatting
func frmt(fr, root, dir, name, group string, t time.Time) string {
	bctx := format.NewContextBuilder()
	bctx.AddString("dir", dir)
	bctx.AddString("name", name)
	bctx.AddString("group", group)
	bctx.AddTime("time", t)
	ctx, err := bctx.BuildContext()
	if err != nil {
		panic(err)
	}
	res, err := format.Format(fr, ctx)
	if err != nil {
		panic(err)
	}
	return filepath.Clean(filepath.Join(root, res))
}

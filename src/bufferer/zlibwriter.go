package bufferer

import (
	"github.com/datadog/czlib"
)

// ZLIBWriter hides switching underlying  after logratition
type ZLIBWriter struct {
	w       *czlib.Writer
	factory func() *czlib.Writer
}

// NewZLIBWriter constructor
func NewZLIBWriter(factory func() *czlib.Writer) *ZLIBWriter {
	w := factory()
	res := &ZLIBWriter{
		w:       w,
		factory: factory,
	}
	return res
}

// Write implementation
func (w *ZLIBWriter) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

// Flush implementation
func (w *ZLIBWriter) Flush() error {
	return w.w.Flush()
}

// Close implementation
func (w *ZLIBWriter) Close() error {
	if err := w.w.Close(); err != nil {
		return err
	}
	w.w = w.factory()
	return nil
}

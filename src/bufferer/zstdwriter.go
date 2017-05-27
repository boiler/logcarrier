package bufferer

import "github.com/Datadog/zstd"

// ZSTDWriter hides switching underlying  after logratition
type ZSTDWriter struct {
	w       *zstd.Writer
	factory func() *zstd.Writer
}

// NewZSTDWriter constructor
func NewZSTDWriter(factory func() *zstd.Writer) *ZSTDWriter {
	w := factory()
	res := &ZSTDWriter{
		w:       w,
		factory: factory,
	}
	return res
}

// Write implementation ...
func (w *ZSTDWriter) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

// Close implementation
func (w *ZSTDWriter) Close() error {
	if err := w.w.Close(); err != nil {
		return err
	}
	w.w = w.factory()
	return nil
}

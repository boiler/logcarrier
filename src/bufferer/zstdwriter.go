package bufferer

import "github.com/Datadog/zstd"

// ZSTDWriter hides switching underlying  after logratition
type ZSTDWriter struct {
	w            *zstd.Writer
	factory      func() *zstd.Writer
	writeCounter int
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
	if len(p) > 0 {
		w.writeCounter++
	}
	return w.w.Write(p)
}

// Close implementation
func (w *ZSTDWriter) Close() error {
	if w.writeCounter == 0 {
		return nil
	}
	if err := w.w.Close(); err != nil {
		return err
	}
	w.w = w.factory()
	w.writeCounter = 0
	return nil
}

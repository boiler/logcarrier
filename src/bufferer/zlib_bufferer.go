package bufferer

import (
	"fileio"
	"frameio"
	"logio"
)

// ZLIBBufferer ...
type ZLIBBufferer struct {
	l *logio.Writer
	c *ZLIBWriter
	f *frameio.Writer
	d *fileio.File
}

// NewZLIBBufferer constructor
func NewZLIBBufferer(l *logio.Writer, c *ZLIBWriter, f *frameio.Writer, d *fileio.File) *ZLIBBufferer {
	res := &ZLIBBufferer{
		l: l,
		c: c,
		f: f,
		d: d,
	}
	return res
}

// Write implementation
func (b *ZLIBBufferer) Write(p []byte) (n int, err error) {
	return b.l.Write(p)
}

// Close implementation
func (b *ZLIBBufferer) Close() error {
	if err := b.l.Flush(); err != nil {
		return err
	}
	if err := b.c.Close(); err != nil {
		return err
	}
	if err := b.f.Flush(); err != nil {
		return err
	}
	if err := b.d.Close(); err != nil {
		return err
	}
	return nil
}

// Flush implementation
func (b *ZLIBBufferer) Flush() error {
	if b.l.WorthFlushing() {
		if err := b.l.Flush(); err != nil {
			return err
		}
		if err := b.c.Flush(); err != nil {
			return err
		}
	}
	if b.f.WorthFlushing() {
		if err := b.f.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// Logrotate implementation
func (b *ZLIBBufferer) Logrotate(newpath string) error {
	return b.d.Logrotate(newpath)
}

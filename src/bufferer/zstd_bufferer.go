package bufferer

import (
	"bytes"
	"fileio"
	"frameio"
	"logio"
	"sync"
)

// ZSTDBufferer ...
type ZSTDBufferer struct {
	l *logio.Writer
	c *ZSTDWriter
	f *frameio.Writer
	d *fileio.File
	p *sync.Pool
}

// NewZSTDBufferer constructor
func NewZSTDBufferer(l *logio.Writer, c *ZSTDWriter, f *frameio.Writer, d *fileio.File) *ZSTDBufferer {
	res := &ZSTDBufferer{
		l: l,
		c: c,
		f: f,
		d: d,
	}
	return res
}

// Write implementation
func (b *ZSTDBufferer) Write(p []byte) (n int, err error) {
	return b.l.Write(p)
}

// Close implementation
func (b *ZSTDBufferer) Close() error {
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
func (b *ZSTDBufferer) Flush() error {
	if b.l.WorthFlushing() {
		if err := b.l.Flush(); err != nil {
			return err
		}
	}
	if b.f.WorthFlushing() {
		if err := b.c.Close(); err != nil {
			return err
		}
		if err := b.f.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// Logrotate implementation
func (b *ZSTDBufferer) Logrotate(newpath string) error {
	return b.d.Logrotate(newpath)
}

// DumpState implementation
func (b *ZSTDBufferer) DumpState() (*bytes.Buffer, error) {
	panic("not implemented")
}

// RestoreState implementation
func (b *ZSTDBufferer) RestoreState(*bytes.Buffer) error {
	panic("not implemented")
}

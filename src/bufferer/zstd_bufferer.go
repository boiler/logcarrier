package bufferer

import (
	"bindec"
	"binenc"
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

// PostWrite implementation
func (b *ZSTDBufferer) PostWrite() error {
	if b.l.OvergrownExtra(nil) {
		return b.l.Flush()
	}
	return nil
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
func (b *ZSTDBufferer) Logrotate() error {
	return b.d.Logrotate()

}

// DumpState implementation
func (b *ZSTDBufferer) DumpState(enc *binenc.Encoder, dest *bytes.Buffer) {
	b.l.DumpState(enc, dest)
}

// RestoreState implementation
func (b *ZSTDBufferer) RestoreState(src *bindec.Decoder) {
	b.l.RestoreState(src)
}

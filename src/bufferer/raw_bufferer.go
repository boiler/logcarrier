package bufferer

import (
	"bindec"
	"binenc"
	"bytes"
	"fileio"
	"logio"
)

// RawBufferer ...
type RawBufferer struct {
	l *logio.Writer
	d *fileio.File
}

// NewRawBufferer constructor
func NewRawBufferer(l *logio.Writer, d *fileio.File) *RawBufferer {
	return &RawBufferer{
		l: l,
		d: d,
	}
}

// Write implementation
func (b *RawBufferer) Write(p []byte) (n int, err error) {
	return b.l.Write(p)
}

// PostWrite implementation
func (b *RawBufferer) PostWrite() error {
	return b.l.Flush()
}

// Close implementation
func (b *RawBufferer) Close() error {
	if err := b.l.Flush(); err != nil {
		return err
	}
	if err := b.d.Close(); err != nil {
		return err
	}
	return nil
}

// Flush implementation
func (b *RawBufferer) Flush() error {
	if b.l.WorthFlushing() {
		if err := b.l.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// Logrotate implementation
func (b *RawBufferer) Logrotate() error {
	return b.d.Logrotate()
}

// DumpState implementation
func (b *RawBufferer) DumpState(enc *binenc.Encoder, dest *bytes.Buffer) {
	b.l.DumpState(enc, dest)
	b.d.DumpState(enc, dest)
}

// RestoreState implementation
func (b *RawBufferer) RestoreState(src *bindec.Decoder) {
	b.l.RestoreState(src)
	b.l.RestoreState(src)
}

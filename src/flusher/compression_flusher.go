package flusher

import (
	"frameio"
	"logio"
)

// CompressFlusher is a tool for flushing io.Writer's chain of compression
type CompressFlusher struct {
	bufferizer *logio.Writer
	framer     *frameio.Writer
}

// NewCompressFlusher constructor. Represents a flushing procedure chain that is required for compression that is line aware:
// 1) You need to bufferize data in order to achieve a lower amount of expensive foreign calls
// 2) Bufferizing is also needed to keeping lines unbroken
// 3) Framing buffering is needed to lower an amount of disk writes.
func NewCompressFlusher(b *logio.Writer, f *frameio.Writer) *CompressFlusher {
	return &CompressFlusher{
		bufferizer: b,
		framer:     f,
	}
}

// Flush implementation.
// Two flushing procedures are needed:
// 1) First one bufferizer need to send rest of the data into compressor
// 2) Need to dump compressed data that was received from compressor
func (cf *CompressFlusher) Flush() error {
	if err := cf.bufferizer.Flush(); err != nil {
		return err
	}
	if err := cf.framer.Flush(); err != nil {
		return err
	}
	return nil
}

// WorthFlushing implementation
func (cf *CompressFlusher) WorthFlushing() bool {
	return cf.bufferizer.WorthFlushing() || cf.framer.WorthFlushing()
}

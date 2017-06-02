/*
package frameio
Bufferized writer that keeps integrity of blocks produced by data compressors.
This is important, since decomporessors usually work upon the stream of data splitted
into frames, where the frame is a solid chunk that is required to be a single piece to be
decompressed.
*/

package frameio

import (
	"bytes"
	"io"
)

const (
	defaultBufferSize = 128 * 1024 * 1024
)

// Writer is a bufferized writer which takes care of comporession frames integrity, i.e.
// it only splits data at frame bounds, not at frame internals.
type Writer struct {
	bufsize int
	writer  io.Writer
	buffer  *bytes.Buffer

	flushCounter     int
	frameInsert      int
	prevFlushCounter int

	worthFlushing bool
}

// NewWriter constructs writer whose buffer has the default size
func NewWriter(writer io.Writer) *Writer {
	return NewWriterSize(writer, defaultBufferSize)
}

// NewWriterSize return a new writer whose buffer has at least specified size
func NewWriterSize(writer io.Writer, size int) *Writer {
	res := &Writer{
		bufsize:       size,
		writer:        writer,
		buffer:        &bytes.Buffer{},
		worthFlushing: true,
	}
	res.buffer.Grow(size)
	return res
}

// Flush flushes all buffered data
func (w *Writer) Flush() error {
	if w.buffer.Len() > 0 {
		w.flushCounter = w.frameInsert
		if _, err := w.buffer.WriteTo(w.writer); err != nil {
			return err
		}
	}
	w.buffer.Reset()
	return nil
}

// Write writes the content of data into the buffer
func (w *Writer) Write(data []byte) (nn int, err error) {
	if w.buffer.Len() > 0 && w.buffer.Len()+len(data) > w.bufsize {
		w.worthFlushing = false
		err = w.Flush()
		if err != nil {
			return
		}
	}
	if len(data) > w.bufsize {
		nn, err = w.writer.Write(data)
		return
	}
	w.frameInsert++
	nn, err = w.buffer.Write(data)
	return
}

// WorthFlushing checks if any write was done after the last check
func (w *Writer) WorthFlushing() bool {
	res := w.worthFlushing && w.frameInsert != w.prevFlushCounter && w.prevFlushCounter == w.flushCounter
	w.prevFlushCounter = w.flushCounter
	w.worthFlushing = true
	return res
}

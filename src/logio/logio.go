/*
package logio
bufferized writers (and readers?) with log line integrity in mind
*/

package logio

import (
	"bindec"
	"binenc"
	"bytes"
	"io"
)

const (
	defaultBufferSize = 128 * 1024 * 1024
)

// Writer is a bufferized writer which takes care of line integrity, i.e. the writer beneath
// will take full lines only
//
// w := NewTextWriter(os.Stdout)
// written, _ := w.Write([]byte("1\n2\n3\n456")
// fmt.Println("Bytes written:", written)
//
// Output:
// 1
// 2
// 3
// Bytes written: 6
type Writer struct {
	bufsize  int
	writer   io.Writer
	buffer   *bytes.Buffer // so called "commited" data, i.e. finalized lines
	linebuf  *bytes.Buffer // a chunk of data at the end which hasn't been followed by \n yet
	finished bool          // finished = true if the previous buffer had \n at the end

	// Statistics data
	lineCount      int
	savedLineCount int
	prevLineCount  int

	worthFlushing bool
}

// NewWriter returns a new writer whose buffer has the default size
func NewWriter(writer io.Writer) *Writer {
	return NewWriterSize(writer, defaultBufferSize)
}

// NewWriterSize returns a new writer whose buffer has at least specified size
func NewWriterSize(writer io.Writer, size int) *Writer {
	res := &Writer{
		bufsize:       size,
		writer:        writer,
		buffer:        &bytes.Buffer{},
		linebuf:       &bytes.Buffer{},
		finished:      true,
		worthFlushing: true,
	}
	if size > 0 {
		res.buffer.Grow(size)
	}
	res.linebuf.Grow(8192)
	return res
}

// Flush flushes all full lines to the underlying io.Writer
func (w *Writer) Flush() error {
	if w.buffer.Len() > 0 {
		if _, err := w.buffer.WriteTo(w.writer); err != nil {
			return err
		}
	}
	w.buffer.Reset()
	w.savedLineCount = w.lineCount
	return nil
}

// FlushAll flush any buffered data to the underlying io.Writer
func (w *Writer) FlushAll() error {
	if err := w.Flush(); err != nil {
		return err
	}
	if w.linebuf.Len() > 0 {
		_, err := io.Copy(w.writer, w.linebuf)
		if err != nil {
			return err
		}
	}
	w.linebuf.Reset()
	return nil
}

// Write writes the content of data into the buffer.
func (w *Writer) Write(data []byte) (nn int, err error) {
	for len(data) > 0 {
		pos := bytes.IndexByte(data, '\n')
		if pos < 0 {
			n, err := w.linebuf.Write(data)
			nn += n
			if err != nil {
				return nn, err
			}
			w.finished = false
			return nn, err
		}

		line := data[:pos+1]
		if !w.finished {
			_, err = w.linebuf.Write(line)
			if err != nil {
				return nn, err
			}
			line = w.linebuf.Bytes()
			w.linebuf.Reset()
		}
		if w.bufsize > 0 {
			if w.buffer.Len()+len(line) > w.bufsize {
				w.worthFlushing = false
				err = w.Flush()
				if err != nil {
					return nn, err
				}
			}
		}
		_, err = w.buffer.Write(line)
		if err != nil {
			return nn, err
		}
		nn += pos + 1
		data = data[pos+1:]
		w.lineCount++
		w.finished = true
	}
	return
}

// LinesBuffered returns how many lines are in buffer now
func (w *Writer) LinesBuffered() int {
	return w.lineCount - w.savedLineCount
}

// LinesWritten returns how many lines were written to the underlying io.Writer
func (w *Writer) LinesWritten() int {
	return w.savedLineCount
}

// WorthFlushing checks if any write was done after the last check
func (w *Writer) WorthFlushing() bool {
	res := w.worthFlushing && w.savedLineCount != w.lineCount && w.savedLineCount == w.prevLineCount
	w.prevLineCount = w.savedLineCount
	w.worthFlushing = true
	return res
}

// DumpState ...
func (w *Writer) DumpState(enc *binenc.BinaryEncoder, dest *bytes.Buffer) {
	dest.Write(enc.Uint32(uint32(w.bufsize)))
	dest.Write(enc.Uint32(uint32(w.buffer.Len())))
	dest.Write(w.buffer.Bytes())
	dest.Write(enc.Uint32(uint32(w.linebuf.Len())))
	dest.Write(w.linebuf.Bytes())
	dest.Write(enc.Bool(w.finished))
	dest.Write(enc.Uint32(uint32(w.lineCount)))
	dest.Write(enc.Uint32(uint32(w.savedLineCount)))
	dest.Write(enc.Uint32(uint32(w.prevLineCount)))
	dest.Write(enc.Bool(w.worthFlushing))
}

// RestoreState ...
func (w *Writer) RestoreState(src *bindec.ResponseReader) {
	bufsize, ok := src.Uint32()
	if !ok {
		panic("Cannot restore bufsize")
	}
	buflen, ok := src.Uint32()
	if !ok {
		panic("Cannot restore buffer length")
	}
	buffer, ok := src.Bytes(int(buflen))
	if !ok {
		panic("Cannot restore a buffer")
	}
	linebuflen, ok := src.Uint32()
	if !ok {
		panic("Cannot restore line buffer length")
	}
	linebuf, ok := src.Bytes(int(linebuflen))
	if !ok {
		panic("Cannot restore line buffer")
	}
	finished, ok := src.Bool()
	if !ok {
		panic("Cannot restore finished state")
	}
	linecount, ok := src.Uint32()
	if !ok {
		panic("Cannot restore line count")
	}
	savedlinecount, ok := src.Uint32()
	if !ok {
		panic("Cannot restore saved line count")
	}
	prevlinecount, ok := src.Uint32()
	if !ok {
		panic("Cannot restore prev(ious) line count")
	}
	worthflushing, ok := src.Bool()
	if !ok {
		panic("Cannot restore worthflushing state")
	}

	w.bufsize = int(bufsize)
	w.buffer.Reset()
	w.buffer.Write(buffer)
	w.linebuf.Reset()
	w.linebuf.Write(linebuf)
	w.finished = finished
	w.lineCount = int(linecount)
	w.savedLineCount = int(savedlinecount)
	w.prevLineCount = int(prevlinecount)
	w.worthFlushing = worthflushing
}

/*
package logio
bufferized writers (and readers?) with log line integrity in mind
*/

package logio

import (
	"bytes"
	"io"
)

const (
	defaultBufferSize = 8192
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
}

// NewWriter returns a new reader whose buffer has the default size
func NewWriter(writer io.Writer) *Writer {
	return NewWriterSize(writer, defaultBufferSize)
}

// NewWriterSize returns a new reader whose buffer has at least specified size
func NewWriterSize(writer io.Writer, size int) *Writer {
	return &Writer{
		bufsize:  size,
		writer:   writer,
		buffer:   &bytes.Buffer{},
		linebuf:  &bytes.Buffer{},
		finished: true,
	}
}

// Flush flushes all full lines to the underlying io.Writer
func (w *Writer) Flush() error {
	if w.buffer.Len() > 0 {
		_, err := io.Copy(w.writer, w.buffer)
		if err != nil {
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

// Write writes the content of p into the buffer.
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
		if w.buffer.Len()+len(line) > w.bufsize {
			err = w.Flush()
			if err != nil {
				return nn, err
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

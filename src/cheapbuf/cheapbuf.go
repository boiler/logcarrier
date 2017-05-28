package cheapbuf

import (
	"bytes"
	"io"
)

const (
	defaultBufferSize = 8192
)

// Reader is a "cheap" bufferized reader that can be switched over underlying
// io.Reader. All responsibility of data integrity is on user side though.
type Reader struct {
	buf []byte
	off int
	cap int

	reader io.Reader
}

// NewReader constructor
func NewReader() *Reader {
	return NewReaderSize(defaultBufferSize)
}

// NewReaderSize constructor
func NewReaderSize(size int) *Reader {
	return &Reader{
		buf:    make([]byte, size),
		reader: nil,
	}
}

// SetReader sets underlying reader
func (r *Reader) SetReader(reader io.Reader) {
	r.reader = reader
}

func (r *Reader) fill() (err error) {
	r.cap, err = r.reader.Read(r.buf)
	r.off = 0
	return
}

// Read implementation for io.Reader
// Warning:
func (r *Reader) Read(p []byte) (n int, err error) {
	if r.off >= r.cap {
		if err = r.fill(); err != nil {
			return
		}
	}
	n = len(p)
	if n > r.cap-r.off {
		n = r.cap - r.off
	}
	copy(p, r.buf[r.off:r.off+n])
	r.off += n
	return
}

// ReadSlice reads a slice from underlying buffer to the delim character
// or to the end if not found
func (r *Reader) ReadSlice(delim byte) (line []byte, err error) {
	if r.off >= r.cap {
		if err = r.fill(); err != nil {
			return
		}
	}
	if i := bytes.IndexByte(r.buf[r.off:r.cap], delim); i >= 0 {
		line = r.buf[r.off : r.off+i+1]
		r.off += i + 1
	} else {
		line = r.buf[r.off:r.cap]
		r.off = r.cap
	}
	return
}

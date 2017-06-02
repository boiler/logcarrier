package cheapbuf

import (
	"bytes"
	"io"
)

// Scanner for scanning lines, with an ability to switch underlying io.Reader
type Scanner struct {
	reader   *Reader
	line     []byte
	buf      *bytes.Buffer
	err      error
	finished bool
}

// NewScanner constructor
func NewScanner(reader *Reader) *Scanner {
	return &Scanner{
		reader:   reader,
		line:     nil,
		buf:      &bytes.Buffer{},
		finished: false,
	}
}

// SetReader sets underlying reader
func (s *Scanner) SetReader(r io.Reader) {
	s.finished = false
	s.reader.SetReader(r)
}

// Scan bool
func (s *Scanner) Scan() bool {
	if s.finished {
		return false
	}
	s.line, s.err = s.reader.ReadSlice('\n')
	if s.err != nil {
		if s.err == io.EOF {
			s.err = nil
			s.finished = true
			return len(s.line) > 0
		}
		return false
	}
	if len(s.line) > 0 && s.line[len(s.line)-1] == '\n' {
		return true
	}
	s.buf.Reset()
	_, _ = s.buf.Write(s.line)
	for {
		line, err := s.reader.ReadSlice('\n')
		if err != nil {
			if err == io.EOF {
				s.finished = true
				_, _ = s.buf.Write(line)
				break
			}
			s.err = err
			return false
		}
		if len(line) > 0 && line[len(line)-1] == '\n' {
			_, _ = s.buf.Write(line)
			s.line = s.buf.Bytes()
			return true
		}
		_, _ = s.buf.Write(line)
	}
	return false
}

// Bytes returns bytes of the line
func (s *Scanner) Bytes() []byte {
	return s.line
}

// Err ...
func (s *Scanner) Err() error {
	return s.err
}

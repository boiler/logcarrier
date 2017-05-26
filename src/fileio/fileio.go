/*
package fileio
*/

package fileio

import "os"

// File object that is steady against rotating.
type File struct {
	flags int
	mode  os.FileMode
	fpath string

	file *os.File
}

// Open opens a file with default flags and mode
func Open(fpath string) (*File, error) {
	return OpenFile(fpath, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.FileMode(0644))
}

// OpenFile with given flags and mode
func OpenFile(fpath string, flags int, mode os.FileMode) (*File, error) {
	file, err := os.OpenFile(fpath, flags, mode)
	if err != nil {
		return nil, err
	}
	return &File{
		flags: flags,
		mode:  mode,
		fpath: fpath,

		file: file,
	}, nil
}

// Write ...
func (f *File) Write(p []byte) (n int, err error) {
	if f.file == nil {
		if f.file, err = os.OpenFile(f.fpath, f.flags, f.mode); err != nil {
			return
		}
	}
	return f.file.Write(p)
}

// Close ...
func (f *File) Close() error {
	if err := f.file.Close(); err != nil {
		return err
	}
	f.file = nil
	return nil
}

/*
package fileio
*/

package fileio

import (
	"fmt"
	"os"
	"utils"
)

// File object that is steady against rotating.
type File struct {
	flags int
	mode  os.FileMode
	root  Root

	fpath string

	file *os.File
}

// Open opens a file with default flags and mode
func Open(root Root, fpath string) (*File, error) {
	return OpenFile(root, fpath, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.FileMode(0644))
}

// OpenFile with given flags and mode
func OpenFile(root Root, fpath string, flags int, mode os.FileMode) (*File, error) {
	file, err := root.OpenFile(fpath, flags, mode)
	if err != nil {
		return nil, err
	}
	return &File{
		flags: flags,
		mode:  mode,
		fpath: fpath,
		root:  root,

		file: file,
	}, nil
}

// Write ...
func (f *File) Write(p []byte) (n int, err error) {
	if f.file == nil {
		if f.file, err = f.root.OpenFile(f.fpath, f.flags, f.mode); err != nil {
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

// Logrotate ...
func (f *File) Logrotate(newpath string) error {
	if f.file != nil {
		panic(fmt.Errorf("File `%s` had to be closed before the log rotation into `%s`", newpath, f.fpath))
	}
	if !utils.PathExists(f.root.Path(f.fpath)) {
		return fmt.Errorf("Can't rename file %s: file not exists", f.fpath)
	}
	if utils.PathExists(f.root.Path(newpath)) {
		return fmt.Errorf("Can't rename file %s => %s: file exists", f.fpath, newpath)
	}
	err := os.Rename(f.root.Path(f.fpath), f.root.Path(newpath))
	return err
}

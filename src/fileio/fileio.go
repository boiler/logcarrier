/*
package fileio
*/

package fileio

import (
	"bindec"
	"binenc"
	"bytes"
	"fmt"
	"logging"
	"os"
	"path/filepath"
	"paths"
	"time"
	"utils"
)

const (
	tempLinkFuse = "21312321321423dshjcjbhquyergyuXrey91=123"
)

// File object that is steady against rotating.
type File struct {
	namegen *paths.Files
	dirmode os.FileMode

	fname string
	link  string

	dir   string
	name  string
	group string

	file       *os.File
	time       time.Time
	writeCount int
}

// Open File constructor
func Open(dir, name, group string, namegen *paths.Files, dirmode os.FileMode) (*File, error) {
	file := &File{
		namegen: namegen,
		dirmode: dirmode,

		dir:   dir,
		name:  name,
		group: group,
	}
	err := file.open()
	return file, err
}

func (f *File) open() (err error) {
	t := time.Now()

	fname := f.namegen.Name(f.dir, f.name, f.group, t)
	dirpath, _ := filepath.Split(fname)
	if err = os.MkdirAll(dirpath, f.dirmode); err != nil {
		return err
	}
	file, err := os.OpenFile(fname, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.FileMode(0644))
	if err != nil {
		return
	}

	lname := f.namegen.Link(f.dir, f.name, f.group, t)
	if len(lname) == 0 {
		return
	}

	dirpath, _ = filepath.Split(lname)
	if err = os.MkdirAll(dirpath, f.dirmode); err != nil {
	}
	if utils.PathExists(lname) {
		_, err := os.Readlink(lname)
		if err != nil {
			return fmt.Errorf("File `%s` exists and it is not a link", lname)
		}
	}

	i := 0
	for i = 0; i < len(fname) && i < len(lname) && fname[i] == lname[i]; i++ {
	}
	target := fname[i:]
	if err = os.Symlink(target, lname+tempLinkFuse); err != nil {
		_ = file.Close()
		return
	}
	if err = os.Rename(lname+tempLinkFuse, lname); err != nil {
		_ = file.Close()
		return
	}

	f.file = file
	f.fname = fname
	f.link = lname
	f.time = t
	return
}

// Write ...
func (f *File) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}
	if f.file == nil {
		if err = f.open(); err != nil {
			return
		}
	}
	f.writeCount++
	return f.file.Write(p)
}

// Close ...
func (f *File) Close() error {
	if f.file == nil {
		return nil
	}
	if err := f.file.Close(); err != nil {
		return err
	}
	f.file = nil
	return nil
}

// Logrotate ...
func (f *File) Logrotate(dir, name, group string) error {
	if f.writeCount == 0 {
		logging.Info("No data collected in %s, omitting log rotation", f.fname)
		return nil
	}
	f.writeCount = 0
	if f.file != nil {
		panic(fmt.Errorf("File must be closed before the log rotation `%s`", f.fname))
	}
	return nil
}

// DumpState ...
func (f *File) DumpState(enc *binenc.Encoder, dest *bytes.Buffer) {
	if f.file == nil {
		if err := f.open(); err != nil {
			panic(err)
		}
	}
	pos, err := f.file.Seek(0, os.SEEK_CUR)
	if err != nil {
		panic(err)
	}
	dest.Write(enc.Int64(pos))
}

// RestoreState ...
func (f *File) RestoreState(src *bindec.Decoder) {
	pos, ok := src.Int64()
	if !ok {
		panic("Cannot restore position in the file")
	}
	err := f.file.Truncate(pos)
	if err != nil {
		panic(fmt.Errorf("Cannot restore position in the file: %s", err))
	}
}

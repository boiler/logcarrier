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

// File object that is steady against rotating.
type File struct {
	namegen paths.Paths
	linkgen paths.Paths
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
func Open(dir, name, group string, namegen, linkgen paths.Paths, dirmode os.FileMode) (*File, error) {
	file := &File{
		namegen: namegen,
		linkgen: linkgen,
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

	lname := f.linkgen.Name(f.dir, f.name, f.group, t)
	if len(lname) == 0 {
		return
	}

	dirpath, _ = filepath.Split(lname)
	if err = os.MkdirAll(dirpath, f.dirmode); err != nil {
	}
	if utils.PathExists(lname) {
		dest, err := os.Readlink(lname)
		if err != nil {
			return fmt.Errorf("File `%s` exists and it is not a link", lname)
		}
		if dest != fname {
			return fmt.Errorf("Link `%s` exists but it does not refer to `%s`", lname, fname)
		}
	} else {
		if err = os.Symlink(fname, lname); err != nil {
			_ = file.Close()
			return
		}
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
	if err := f.file.Close(); err != nil {
		return err
	}
	f.file = nil
	return nil
}

// Logrotate ...
func (f *File) Logrotate(dir, name, group string) error {
	rotname := f.namegen.Rotation(dir, name, group, f.time)
	if f.writeCount == 0 {
		logging.Info("No data collected in %s, omitting log rotation", rotname)
		return nil
	}
	f.writeCount = 0
	if f.file != nil {
		panic(fmt.Errorf("File must be closed before the log rotation `%s`=>`%s`", f.fname, rotname))
	}
	if !utils.PathExists(f.fname) {
		return fmt.Errorf("Can't rename file %s: file not exists", f.fname)
	}
	if utils.PathExists(rotname) {
		return fmt.Errorf("Can't rename file %s => %s: file exists", f.fname, rotname)
	}
	if len(f.link) > 0 {
		if err := os.Remove(f.link); err != nil {
			return fmt.Errorf("Can't remove symlink %s => %s: %s", f.link, f.fname, err)
		}
	}
	if err := os.Rename(f.fname, rotname); err != nil {
		return fmt.Errorf("Can't rename file %s => %s: %s", f.fname, rotname, err)
	}
	rotlink := f.linkgen.Rotation(f.dir, f.name, f.group, f.time)
	if len(rotlink) > 0 {
		if err := os.Symlink(rotname, rotlink); err != nil {
			return fmt.Errorf("Can't create symlink %s => %s: %s", rotlink, rotname, err)
		}
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

package main

import (
	"bufferer"
	"fmt"
	"logging"
	"path/filepath"
	"sync"
	"time"

	"github.com/LK4D4/trylock"
)

// Buf access
type Buf struct {
	Dir   string
	Name  string
	Group string
	Key   string
	Lock  *trylock.Mutex
	Buf   bufferer.Bufferer
}

// FileOp suit to work with files (append data to files, logrotate them)
type FileOp struct {
	items     map[string]*Buf
	itemsLock *sync.Mutex
	factory   func(string, string, string) (bufferer.Bufferer, error) // Generates bufferer for a given key

	ticker      *time.Ticker
	stopChannel chan int
	wg          *sync.WaitGroup
}

// NewFileOp generates file service
//   factory creates bufferer object
//   ticker is used to generate
func NewFileOp(factory func(string, string, string) (bufferer.Bufferer, error), ticker *time.Ticker) *FileOp {
	res := &FileOp{
		items:       make(map[string]*Buf),
		itemsLock:   &sync.Mutex{},
		factory:     factory,
		ticker:      ticker,
		stopChannel: make(chan int),
		wg:          &sync.WaitGroup{},
	}

	return res
}

func fileKey(dir, name, group string) string {
	return filepath.Clean(filepath.Join(dir, name, group))
}

// GetFile retrieves Buf
func (f *FileOp) GetFile(dir, name, group string) (res *Buf, err error) {
	f.itemsLock.Lock()
	key := fileKey(dir, name, group)
	buf, ok := f.items[key]
	if !ok {
		b, err := f.factory(dir, name, group)
		if err != nil {
			f.itemsLock.Unlock()
			return res, err
		}
		buf = &Buf{
			Dir:   dir,
			Name:  name,
			Group: group,
			Key:   key,
			Lock:  &trylock.Mutex{},
			Buf:   b,
		}
		f.items[key] = buf
	}
	f.itemsLock.Unlock()
	return buf, nil
}

// Logrotate obviously logrotates file
func (f *FileOp) Logrotate(dir, name, group string) (err error) {
	f.itemsLock.Lock()
	key := fileKey(dir, name, group)
	buf, ok := f.items[key]
	f.itemsLock.Unlock()
	if !ok {
		return fmt.Errorf("file `%s` not found", name)
	}
	buf.Lock.Lock()
	if err := buf.Buf.Close(); err != nil {
		goto exit
	}
	err = buf.Buf.Logrotate(dir, name, group)

exit:
	buf.Lock.Unlock()
	return err
}

// Join wait for the background worker to stop
func (f *FileOp) Join() {
	f.stopChannel <- 0
	f.stopChannel <- 0
	f.wg.Wait()
}

// FlushPeriodic periodically flushes ...
func (f *FileOp) FlushPeriodic() {
	go func() {
		f.wg.Add(1)
		logging.Info("FLUSHER: started")

		buf := make([]*Buf, 4096)

		for {
			select {
			case t := <-f.ticker.C:
				flushed := 0
				flushErrors := 0
				wereLocked := 0

				buf = buf[:0]
				f.itemsLock.Lock()
				for _, v := range f.items {
					buf = append(buf, v)
				}
				f.itemsLock.Unlock()

				logging.Info("FLUSHER: flushing %d items", len(buf))
				for _, v := range buf {
					locked := v.Lock.TryLock()
					if !locked {
						wereLocked++
						continue
					}
					if err := v.Buf.Flush(); err != nil {
						logging.Error("FLUSHER: error flushing \033[1m%s\033[0m, \033[1m%s\033[0m", v.Name, err)
						flushErrors++
					} else {
						flushed++
					}
					v.Lock.Unlock()
				}
				logging.Info(
					`FLUSHER: flushed: %d, were locked: %d, flushes failed: %d, duration: %s`,
					flushed, wereLocked, flushErrors, time.Now().Sub(t))
			case <-f.stopChannel:
				logging.Info("FLUSHER: was ordered to stop flushing, closing buffers")
				f.itemsLock.Lock()
				for k, v := range f.items {
					if err := v.Buf.Close(); err != nil {
						logging.Error("Failed to close %s: %s", k, err)
					} else {
						logging.Info("Closed %s", k)
					}
				}
				f.itemsLock.Unlock()
				f.wg.Done()
				return
			}
		}
	}()
}

// LogrotatePeriodic periodically logrotates all files under write
func (f *FileOp) LogrotatePeriodic(periodic chan int) {
	go func() {
		f.wg.Add(1)
		logging.Info("LOGROTATER: periodic rotater started")
		buf := make([]*Buf, 4096)
		for {
			select {
			case <-periodic:
				buf := buf[:0]
				f.itemsLock.Lock()
				for _, v := range f.items {
					buf = append(buf, v)
				}
				f.itemsLock.Unlock()
				logging.Info("LOGROTATER: periodic rotation of %d items", len(buf))
				for _, v := range buf {
					if err := f.Logrotate(v.Dir, v.Name, v.Group); err != nil {
						logging.Error("LOGROTATER: error. %s", err)
					}
				}

			case <-f.stopChannel:
				f.wg.Done()
				return
			}
		}
	}()
}

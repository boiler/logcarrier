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
	Name string
	Lock *trylock.Mutex
	Buf  bufferer.Bufferer
}

// FileOp suit to work with files (append data to files, logrotate them)
type FileOp struct {
	items     map[string]Buf
	itemsLock *sync.Mutex
	factory   func(string, string, string) (bufferer.Bufferer, error) // Generates bufferer for a given key

	ticker      *time.Ticker
	stopChannel chan int
}

// NewFileOp generates file service
//   factory creates bufferer object
//   ticker is used to generate
func NewFileOp(factory func(string, string, string) (bufferer.Bufferer, error), ticker *time.Ticker) *FileOp {
	res := &FileOp{
		items:       make(map[string]Buf),
		itemsLock:   &sync.Mutex{},
		factory:     factory,
		ticker:      ticker,
		stopChannel: make(chan int),
	}

	go func() {
		logging.Info("FLUSHER: started")

		buf := make([]Buf, 4096)

		for {
			select {
			case t := <-ticker.C:
				flushed := 0
				flushErrors := 0
				wereLocked := 0

				buf = buf[:0]
				res.itemsLock.Lock()
				for _, v := range res.items {
					buf = append(buf, v)
				}
				res.itemsLock.Unlock()

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
					`FLUSHER:
flushed: %d
were locked: %d
flushes failed: %d
duration: %s`,
					flushed, wereLocked, flushErrors, time.Now().Sub(t))
			case <-res.stopChannel:
				logging.Info("FLUSHER: was ordered to stop flushing, closing buffers")
				res.itemsLock.Lock()
				for k, v := range res.items {
					if err := v.Buf.Close(); err != nil {
						logging.Error("Failed to close %s: %s", k, err)
					} else {
						logging.Info("Closed %s", k)
					}
				}
				res.itemsLock.Unlock()
				res.stopChannel <- 0
				return
			}
		}
	}()

	return res
}

func fileKey(dir, name, group string) string {
	return filepath.Clean(filepath.Join(dir, name, group))
}

// GetFile retrieves Buf
func (f *FileOp) GetFile(dir, name, group string) (res Buf, err error) {
	f.itemsLock.Lock()
	key := fileKey(dir, name, group)
	buf, ok := f.items[key]
	if !ok {
		b, err := f.factory(dir, name, group)
		if err != nil {
			f.itemsLock.Unlock()
			return res, err
		}
		buf = Buf{
			Name: key,
			Lock: &trylock.Mutex{},
			Buf:  b,
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
	err = buf.Buf.Logrotate()

exit:
	buf.Lock.Unlock()
	return err
}

// Join wait for the background worker to stop
func (f *FileOp) Join() {
	f.stopChannel <- 0
	<-f.stopChannel
}

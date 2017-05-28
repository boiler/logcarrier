package main

import (
	"bufferer"
	"fmt"
	"logging"
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
	factory   func(string) bufferer.Bufferer // Generates bufferer for a given key

	ticker *time.Ticker

	stop bool
}

// NewFileOp generates file service
//   factory creates bufferer object
//   ticker is used to generate
func NewFileOp(factory func(string) bufferer.Bufferer, ticker *time.Ticker) *FileOp {
	res := &FileOp{
		items:     make(map[string]Buf),
		itemsLock: &sync.Mutex{},
		factory:   factory,
		ticker:    ticker,
		stop:      false,
	}

	go func() {
		logging.Info("FLUSHER: started")

		buf := make([]Buf, 4096)
		for t := range res.ticker.C {
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
		}
		logging.Info("FLUSHER: was ordered to stop flushing, leaving")
	}()

	return res
}

// GetFile retrieves Buf
func (f *FileOp) GetFile(name string) Buf {
	f.itemsLock.Lock()
	buf, ok := f.items[name]
	if !ok {
		b := f.factory(name)
		buf = Buf{
			Name: name,
			Lock: &trylock.Mutex{},
			Buf:  b,
		}
		f.items[name] = buf
	}
	f.itemsLock.Unlock()
	return buf
}

// Logrotate obviously logrotates file
func (f *FileOp) Logrotate(name, newpath string) (err error) {
	f.itemsLock.Lock()
	buf, ok := f.items[name]
	f.itemsLock.Unlock()
	if !ok {
		return fmt.Errorf("file `%s` not found", name)
	}
	buf.Lock.Lock()
	if err := buf.Buf.Close(); err != nil {
		goto exit
	}
	err = buf.Buf.Logrotate(newpath)

exit:
	buf.Lock.Unlock()
	return err
}

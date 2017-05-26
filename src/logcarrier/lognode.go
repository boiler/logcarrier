package main

import (
	"flusher"
	"io"

	"github.com/LK4D4/trylock"
)

// LogNode all operations needed for feeding logs, flushing and log rotating
type LogNode struct {
	writer  io.Writer
	flusher flusher.Flusher
	closer  io.Closer
	locker  *trylock.Mutex
}

// NewLogNode creates a primitve
func NewLogNode(writer io.Writer, flusher flusher.Flusher, closer io.Closer) *LogNode {
	return &LogNode{
		writer:  writer,
		flusher: flusher,
		closer:  closer,
		locker:  &trylock.Mutex{},
	}
}

// Write ...
func (ln *LogNode) Write(p []byte) (n int, err error) {
	return ln.writer.Write(p)
}

// Close ...
func (ln *LogNode) Close() error {
	return ln.closer.Close()
}

// Flush flushes once writers beneath report it is time to
func (ln *LogNode) Flush() error {
	if ln.flusher.WorthFlushing() {
		return ln.flusher.Flush()
	}
	return nil
}

// Lock ...
func (ln *LogNode) Lock() {
	ln.locker.Lock()
}

// Unlock ...
func (ln *LogNode) Unlock() {
	ln.locker.Unlock()
}

// TryLock ...
func (ln *LogNode) TryLock() bool {
	return ln.locker.TryLock()
}

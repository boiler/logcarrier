package main

import (
	"cheapbuf"
	"connio"
	"fmt"
	"logging"
	"net"
	"sync"
	"time"
)

var protocol2Header = []byte("200 READY protocol 2\n")
var protocol1Header = []byte("200 READY\n")
var okMsg = []byte("200 OK\n")

// DumpJob connection with tailed data
type DumpJob struct {
	Name string   // Name of the file
	Conn net.Conn // Socket
	Size int      // Bytes to read (-1 for unknown)
}

// DumpPool spawns workers what read incoming data from tailers
type DumpPool struct {
	netjobs chan DumpJob
	files   *FileOp

	jobsCounter int
	wg          *sync.WaitGroup
	waitTimeout time.Duration
	stopQueue   chan int
}

// NewDumpPool constructor
func NewDumpPool(netjobs chan DumpJob, files *FileOp, timeout time.Duration) *DumpPool {
	return &DumpPool{
		netjobs:     netjobs,
		files:       files,
		stopQueue:   make(chan int),
		waitTimeout: timeout,
		wg:          &sync.WaitGroup{},
	}
}

// Stop command jobs to stop
func (dp *DumpPool) Stop() {
	logging.Info("Stopping dumping jobs")
	for i := 0; i < dp.jobsCounter; i++ {
		dp.stopQueue <- 0
	}
	dp.wg.Wait()
	logging.Info("Done")
}

type worker struct {
	buffer  []byte
	scanner *cheapbuf.Scanner
	conn    *connio.Reader
}

// Spawn spawns worker
func (dp *DumpPool) Spawn() {
	dp.jobsCounter++
	go func() {
		dp.wg.Add(1)
		w := &worker{
			buffer:  make([]byte, 4096),
			scanner: cheapbuf.NewScanner(cheapbuf.NewReaderSize(8129)),
			conn:    connio.NewReader(dp.waitTimeout),
		}
		for {
			select {
			case x := <-dp.netjobs:
				if err := dp.dump(x, w); err != nil {
					logging.Error("DUMPING: %s", err)
				}
				if err := x.Conn.Close(); err != nil {
					logging.Error("DUMPING: cannot close connection to %s: %s", x.Conn.RemoteAddr().String(), err)
				}
			case <-dp.stopQueue:
				dp.wg.Done()
				return
			}
		}
	}()
}

func (dp *DumpPool) dump(x DumpJob, w *worker) (err error) {
	buf, err := dp.files.GetFile(x.Name)
	if err != nil {
		return
	}
	buf.Lock.Lock()
	left := x.Size
	if x.Size > 0 { // Protocol 2
		if _, err := x.Conn.Write(protocol2Header); err != nil {
			return fmt.Errorf("Failed to set a protocol for  %s: %s", x.Name, err)
		}
		w.conn.SetConn(x.Conn)
		for left > 0 {
			read, err := w.conn.Read(w.buffer)
			if err != nil {
				return fmt.Errorf("Error when reading for `%s`: %s", x.Name, err.Error())
			}
			if _, err := buf.Buf.Write(w.buffer[:read]); err != nil {
				return fmt.Errorf("Error writing data for `%s`: %s", x.Name, err.Error())
			}
			left -= read
		}
	} else { // Protocol 1
		if _, err := x.Conn.Write(protocol1Header); err != nil {
			return fmt.Errorf("Failed to set protocol for %s: %s", x.Name, err)

		}
		w.conn.SetConn(x.Conn)
		w.scanner.SetReader(w.conn)
		for w.scanner.Scan() {
			line := w.scanner.Bytes()

			// Check if line is the last one
			if len(line) > 1 {
				if line[0] == '.' {
					ws := true
					for i := 1; i < len(line); i++ {
						if line[i] == '\r' || line[i] == '\n' {
							ws = false
							break
						}
					}
					if ws {
						break
					}
					line = line[1:]
				}
			}

			if _, err := buf.Buf.Write(line); err != nil {
				return fmt.Errorf("Failed to write into buffer for %s: %s", x.Name, err)
			}
		}
	}
	if _, err := x.Conn.Write(okMsg); err != nil {
		return fmt.Errorf("Failed to confirm successful read to the tailer for %s: %s", x.Name, err)
	}
	return nil
}

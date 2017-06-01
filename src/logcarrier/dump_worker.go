package main

import (
	"bindec"
	"binenc"
	"bufferer"
	"bytes"
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
var errMsg = []byte("400 Error\n")

// DumpJob connection with tailed data
type DumpJob struct {
	Dir   string //
	Name  string // Name of the file
	Group string
	Conn  net.Conn // Socket
	Size  int      // Bytes to read (-1 for unknown)
}

// DumpPool spawns workers what read incoming data from tailers
type DumpPool struct {
	netjobs chan DumpJob
	files   *FileOp

	jobsCounter int
	wg          *sync.WaitGroup
	waitTimeout time.Duration
	stopQueue   chan int
	pool        *sync.Pool
}

// NewDumpPool constructor
func NewDumpPool(netjobs chan DumpJob, files *FileOp, timeout time.Duration) *DumpPool {
	return &DumpPool{
		netjobs:     netjobs,
		files:       files,
		stopQueue:   make(chan int),
		waitTimeout: timeout,
		wg:          &sync.WaitGroup{},
		pool: &sync.Pool{
			New: func() interface{} { return &bytes.Buffer{} },
		},
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
		enc := binenc.New()
		dec := bindec.New(nil)
		for {
			select {
			case x := <-dp.netjobs:
				if err := dp.dump(x, w, enc, dec); err != nil {
					logging.Error("DUMPING: %s", err)
				}
				if err := x.Conn.Close(); err != nil {
					logging.Info("DUMPING: closing connection")
					logging.Error("DUMPING: cannot close connection to %s: %s", x.Conn.RemoteAddr().String(), err)
				}
			case <-dp.stopQueue:
				dp.wg.Done()
				return
			}
		}
	}()
}

func (dp *DumpPool) dump(x DumpJob, w *worker, e *binenc.Encoder, d *bindec.Decoder) (err error) {
	buf, err := dp.files.GetFile(x.Dir, x.Name, x.Group)
	if err != nil {
		return
	}
	buf.Lock.Lock()
	dest := dp.pool.Get().(*bytes.Buffer)
	dest.Reset()
	buf.Buf.DumpState(e, dest)
	buf.Counter = 3
	err = dp.communicate(x, w, buf.Buf)
	if err != nil {
		d.SetSource(dest.Bytes())
		buf.Buf.RestoreState(d)
	}
	dp.pool.Put(dest)
	buf.Lock.Unlock()
	if err != nil {
		if _, nerr := x.Conn.Write(errMsg); nerr != nil {
			logging.Error("Failed to signal error to tailer of %s: %s", x.Name, nerr)
		}
		return err
	}
	if _, err := x.Conn.Write(okMsg); err != nil {
		return fmt.Errorf("Failed to confirm successful read to the tailer for %s: %s", x.Name, err)
	}
	return nil
}

func (dp *DumpPool) communicate(x DumpJob, w *worker, buf bufferer.Bufferer) (err error) {
	left := x.Size
	if x.Size > 0 { // Protocol 2
		if _, err = x.Conn.Write(protocol2Header); err != nil {
			return fmt.Errorf("Failed to start protocol 2 for %s: %s", x.Name, err)
		}
		w.conn.SetConn(x.Conn)
		for left > 0 {
			read, err := w.conn.Read(w.buffer)
			if err != nil {
				return fmt.Errorf("Error when reading data (protocol 2) %s: %s", x.Name, err)
			}
			if _, err = buf.Write(w.buffer[:read]); err != nil {
				return fmt.Errorf("Error writing retrieved data for %s: %s", x.Name, err)
			}
			left -= read
		}
	} else { // Protocol 1
		if _, err = x.Conn.Write(protocol1Header); err != nil {
			return fmt.Errorf("Failed to start protocol 1 for %s: %s", x.Name, err)
		}
		w.conn.SetConn(x.Conn)
		w.scanner.SetReader(w.conn)
		for w.scanner.Scan() {
			line := w.scanner.Bytes()

			if len(line) > 1 {
				if line[0] == '.' {
					ws := true
					for i := 1; i < len(line); i++ {
						if line[i] != '\r' && line[i] != '\n' {
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
			if _, err = buf.Write(line); err != nil {
				return fmt.Errorf("Error writing retrieved data for %s: %s", x.Name, err)
			}
		}
	}
	return buf.PostWrite()
}

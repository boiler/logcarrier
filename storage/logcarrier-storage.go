package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"./config"
	"./logging"
)

import _ "net/http/pprof"

var locksCount int32 = 0

type Locks struct {
	sync.RWMutex
	fmap map[string]*sync.RWMutex
}

type Config struct {
	Listen      string        `toml:"listen"`
	ListenDebug string        `toml:"listen_debug"`
	WaitTimeout time.Duration `toml:"wait_timeout"`
	Key         string        `toml:"key"`
	DestDir     string        `toml:"destdir"`
	DestDirMode os.FileMode   `toml:"destdir_mode"`
	LogFile     string        `toml:"logfile"`
}

func newConfig() *Config {
	config := &Config{}
	config.Listen = "0.0.0.0:1466"
	config.ListenDebug = ""
	config.WaitTimeout = 60
	config.Key = "key"
	config.DestDir = "./logs"
	config.DestDirMode = 0755
	config.LogFile = ""
	return config
}

// PathExists checks that path exists on filesystem
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func main() {

	flag.Parse()

	cfg := newConfig()
	if err := config.Parse(cfg); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	if len(cfg.LogFile) > 0 {
		loggingConfig := logging.NewConfig()
		loggingConfig.Logfile = cfg.LogFile
		logging.SetConfig(loggingConfig)
	}

	logging.Info("Started")

	if !PathExists(cfg.DestDir) {
		fmt.Fprintf(os.Stderr, "Error: Directory %v not exists\n", cfg.DestDir)
		os.Exit(1)
	}

	if len(cfg.ListenDebug) > 0 {
		logging.Info("Debug listening on " + cfg.ListenDebug)
		go func() {
			http.ListenAndServe(cfg.ListenDebug, nil)
		}()
	}

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	l, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		logging.Critical("Error listening: %s", err.Error())
		os.Exit(1)
	}

	defer l.Close()
	logging.Info("Listening on " + cfg.Listen)
	acceptConn := true

	locks := &Locks{
		fmap: make(map[string]*sync.RWMutex),
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				logging.Critical("Error accepting: %s", err.Error())
			}
			go handleRequest(conn, cfg, locks)
			if !acceptConn {
				break
			}
		}
	}()

sigLoop:
	for {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			logging.Info("SIGINT received")
			acceptConn = false
			break sigLoop
		case syscall.SIGTERM:
			logging.Info("SIGTERM received")
			acceptConn = false
			break sigLoop
		}
	}

	i := 0
	for {
		lcnt := atomic.LoadInt32(&locksCount)
		if lcnt < 1 {
			break
		}
		if i == 0 {
			logging.Info("Waiting for %d locks", lcnt)
		}
		i++
		if i > 100 {
			i = 0
		}
		time.Sleep(100 * time.Millisecond)
	}

	logging.Info("EXIT")
}

// Handles incoming requests.
func handleRequest(conn net.Conn, cfg *Config, locks *Locks) {
	conn.SetDeadline(time.Now().Add(60 * time.Second))
	defer conn.Close()

	remoteAddr, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return
	}

	lineslc := strings.Fields(string(line))
	if len(lineslc) < 5 {
		return
	}

	acmd := lineslc[0]
	akey := lineslc[1]
	dname := lineslc[3]
	fname := lineslc[4]
	bcnt := 0

	dpath := path.Join(cfg.DestDir, dname)
	fpath := path.Join(dpath, fname)

	fpathAbs, _ := filepath.Abs(fpath)
	cfgDirAbs, _ := filepath.Abs(cfg.DestDir)

	if !strings.HasPrefix(fpathAbs, cfgDirAbs) {
		logging.Error("%s unsecure file path %s => %s", remoteAddr, dname, fpathAbs)
		return
	}

	if acmd == "DATA" {
		if len(lineslc) > 5 { // protocol 2
			i, err := strconv.Atoi(lineslc[5])
			if err == nil {
				bcnt = i
			}
		}
	} else if acmd == "ROTATE" {
		t := time.Now()
		newfname := fmt.Sprintf("%s-%d%02d%02d%02d%02d%02d", fname, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
		if len(lineslc) > 5 {
			newfname = lineslc[5]
		}
		newfpath := path.Join(dpath, newfname)
		newfpathAbs, _ := filepath.Abs(newfpath)
		if !PathExists(fpathAbs) {
			logging.Error("Can't rename file %s: file not exists", fpathAbs)
			conn.Write([]byte("400 Error\n"))
			return
		}
		if PathExists(newfpathAbs) {
			logging.Error("Can't rename file %s => %s: file exists", fpathAbs, newfpathAbs)
			conn.Write([]byte("400 Error\n"))
			return
		}
		if !strings.HasPrefix(newfpathAbs, cfgDirAbs) {
			logging.Error("%s unsecure file path %s => %s", remoteAddr, dname, newfpathAbs)
			conn.Write([]byte("400 Error\n"))
			return
		}
		err := os.Rename(fpathAbs, newfpathAbs)
		if err == nil {
			conn.Write([]byte("200 DONE\n"))
			logging.Info("File rotated %s => %s", fpathAbs, newfpathAbs)
			return
		}
		logging.Error("Can't rename file %s => %s", fpathAbs, newfpathAbs)
		conn.Write([]byte("400 Error\n"))
		return
	} else {
		logging.Error("%s unknown command", remoteAddr)
		return
	}
	if akey != cfg.Key {
		logging.Error("%s wrong key", remoteAddr)
		return
	}

	if !PathExists(dpath) {
		os.MkdirAll(dpath, cfg.DestDirMode)
	}

	locks.Lock()
	if _, ok := locks.fmap[fpath]; !ok {
		locks.fmap[fpath] = new(sync.RWMutex)
	}
	flock := locks.fmap[fpath]
	locks.Unlock()
	atomic.AddInt32(&locksCount, 1)
	flock.Lock()

	const fileflag int = os.O_CREATE | os.O_APPEND | os.O_RDWR
	const filemode os.FileMode = 0644
	f, err := os.OpenFile(fpath, fileflag, filemode)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	defer flock.Unlock()
	defer atomic.AddInt32(&locksCount, -1)

	ok := false
	linesNum := 0
	bytesNum := 0
	bytesNumW := 0
	fpos, _ := f.Seek(0, 2)
	w := bufio.NewWriter(f)

	if bcnt > 0 { // protocol 2
		conn.Write([]byte("200 READY protocol 2\n"))
		brem := bcnt
		for {
			conn.SetDeadline(time.Now().Add(cfg.WaitTimeout * time.Second))
			buf := make([]byte, 1024)
			bn, err := reader.Read(buf)
			if err != nil {
				logging.Error("Can't read socket on %s: %s", fpathAbs, err)
				break
			}
			bnw, err := w.Write(buf[:bn])
			if err != nil {
				logging.Error("Can't write to %s: %s", fpathAbs, err)
				break
			}
			bytesNum += bn
			bytesNumW += bnw
			brem -= bn
			if brem == 0 {
				if bytesNum == bcnt {
					ok = true
				} else {
					logging.Error("Read %d bytes of %d on %s", bytesNum, bcnt, fpathAbs)
				}
				break
			}
		}
	} else { // protocol 1
		conn.Write([]byte("200 READY\n"))
		for {
			conn.SetDeadline(time.Now().Add(cfg.WaitTimeout * time.Second))
			line, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}
			if line[0] == '.' {
				tline := bytes.TrimRight(line, "\n\r")
				if len(tline) == 1 {
					ok = true
					break
				}
				if len(bytes.TrimLeft(tline, ".")) == 0 {
					line = line[1:]
				}
			}
			bn, err := w.Write(line)
			if err != nil {
				logging.Error("Can't write to %s: %s", fpathAbs, err)
				break
			}
			bytesNum += bn
			linesNum++
		}
	}
	if ok {
		w.Flush()
		if bcnt > 0 { // protocol 2
			logging.Info("%s %s/%s %d", remoteAddr, dname, fname, bytesNum)
			if bytesNum != bytesNumW {
				logging.Error("Read bytes are not equals write bytes: %s %s %s", fname, bytesNum, bytesNumW)
			}
		} else {
			logging.Info("%s %s/%s %d %d", remoteAddr, dname, fname, linesNum, bytesNum)
		}
		conn.Write([]byte("200 OK\n"))
	} else {
		f.Truncate(fpos)
		logging.Error("%s %s file truncated", remoteAddr, fpath)
	}
}

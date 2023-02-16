package logging

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/howeyc/fsnotify"
)

var std = NewFileLogger()

func init() {
	logrus.SetFormatter(&TextFormatter{})

	// signal watcher
	signalChan := make(chan os.Signal, 16)
	signal.Notify(signalChan, syscall.SIGHUP)

	go func() {
		for {
			select {
			case <-signalChan:
				err := std.Reopen()
				logrus.Infof("HUP received, reopen log %#v", std.Filename())
				if err != nil {
					logrus.Errorf("Reopen log %#v failed: %#s", std.Filename(), err.Error())
				}
			}
		}
	}()
}

// FileLogger обертка
type FileLogger struct {
	sync.RWMutex
	filename    string
	fd          *os.File
	logger      *logrus.Logger
	watcherDone chan bool
}

// NewFileLogger создает инстанс FileLogger
func NewFileLogger() *FileLogger {
	return &FileLogger{}
}

// Write пишет ошибку в Info уровне в текущий вывод.
// Полезно исключительно для перенаправления вывода стандартного log'а
func (l FileLogger) Write(p []byte) (int, error) {
	l.Lock()
	defer l.Unlock()

	// remove extra newline if any
	if p[len(p)-1] == '\n' {
		p = p[:len(p)-1]
	}

	if l.logger != nil {
		l.logger.Infof("%s", p)
	} else {
		logrus.Infof("%s", p)
	}
	return len(p), nil
}

// Open устанавливает имя файла для логирования и открывает его
func (l *FileLogger) Open(filename string) error {
	l.Lock()
	l.filename = filename
	l.Unlock()

	reopenErr := l.Reopen()
	if l.watcherDone != nil {
		close(l.watcherDone)
	}
	l.watcherDone = make(chan bool)
	l.fsWatch(l.filename, l.watcherDone)

	return reopenErr
}

func (l *FileLogger) fsWatch(filename string, quit chan bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Warningf("fsnotify.NewWatcher(): %s", err)
		return
	}

	if filename == "" {
		return
	}

	subscribe := func() {
		if err := watcher.WatchFlags(filename, fsnotify.FSN_CREATE|fsnotify.FSN_DELETE|fsnotify.FSN_RENAME); err != nil {
			logrus.Warningf("fsnotify.Watcher.Watch(%s): %s", filename, err)
		}
	}

	subscribe()

	go func() {
		defer watcher.Close()

		for {
			select {
			case <-watcher.Event:
				l.Reopen()
				subscribe()

				logrus.Infof("Reopen log %#v by fsnotify event", std.Filename())
				if err != nil {
					logrus.Errorf("Reopen log %#v failed: %#s", std.Filename(), err.Error())
				}

			case <-quit:
				return
			}
		}
	}()
}

// Reopen переоткрывает файл
func (l *FileLogger) Reopen() error {
	l.Lock()
	defer l.Unlock()

	var newFd *os.File
	var err error

	if l.filename != "" {
		newFd, err = os.OpenFile(l.filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

		if err != nil {
			return err
		}
	} else {
		newFd = nil
	}

	oldFd := l.fd
	l.fd = newFd

	var loggerOut io.Writer

	if l.fd != nil {
		loggerOut = l.fd
	} else {
		loggerOut = os.Stderr
	}
	logrus.SetOutput(loggerOut)

	if oldFd != nil {
		oldFd.Close()
	}

	return nil
}

// Filename возвращает текущее имя файла
func (l *FileLogger) Filename() string {
	l.RLock()
	l.RUnlock()
	return l.filename
}

// PrepareFile creates logfile and set it writable for user
func PrepareFile(filename string, owner *user.User) error {
	if filename == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if fd != nil {
		fd.Close()
	}
	if err != nil {
		return err
	}
	if err := os.Chmod(filename, 0644); err != nil {
		return err
	}
	if owner != nil {

		uid, err := strconv.ParseInt(owner.Uid, 10, 0)
		if err != nil {
			return err
		}

		gid, err := strconv.ParseInt(owner.Gid, 10, 0)
		if err != nil {
			return err
		}

		if err := os.Chown(filename, int(uid), int(gid)); err != nil {
			return err
		}
	}

	return nil
}

// func (w *wrapper) setBackend(backend *gologging.LogBackend) {
// 	var format = gologging.MustStringFormatter(
// 		"[%{time:2006-01-02 15:04:05}] %{level:.1s} %{message}",
// 	)

// 	backendFormatter := gologging.NewBackendFormatter(backend, format)
// 	backendLeveled := gologging.AddModuleLevel(backendFormatter)
// 	w.Logger.SetBackend(backendLeveled)
// }

// SetFile перенаправляет логи в файл
func SetFile(filename string) error {
	return std.Open(filename)
}

// SetConfig настраивает логирование по конфигу
func SetConfig(config *Config) error {
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return err
	}
	logrus.SetLevel(level)
	if err := std.Open(config.Logfile); err != nil {
		return err
	}
	log.SetFlags(0)
	log.SetOutput(std)
	return nil
}

type LogInterface interface {
	Log(args ...interface{})
}

type TestingLogger struct {
	t        LogInterface
	oldLevel logrus.Level
	oldFlags int
}

func (l TestingLogger) Write(p []byte) (int, error) {
	end := len(p) - 1
	for p[end] == '\n' {
		end--
	}
	l.t.Log(string(p[:end+1]))
	return len(p), nil
}

func SetTesting(t LogInterface) *TestingLogger {
	proxyLogger := &TestingLogger{
		t:        t,
		oldLevel: logrus.GetLevel(),
		oldFlags: log.Flags(),
	}
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(proxyLogger)
	log.SetOutput(proxyLogger)
	log.SetFlags(0)
	return proxyLogger
}

func UnsetTesting(l *TestingLogger) {
	logrus.SetLevel(l.oldLevel)
	logrus.SetOutput(os.Stderr)
	log.SetOutput(os.Stderr)
	log.SetFlags(l.oldFlags)
}

// Critical logs a message using CRITICAL as log level.
func Critical(format string, args ...interface{}) {
	logrus.Errorf(fmt.Sprintf("C %s", format), args...)
}

// Error logs a message using ERROR as log level.
func Error(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

// Warning logs a message using WARNING as log level.
func Warning(format string, args ...interface{}) {
	logrus.Warningf(format, args...)
}

// Info logs a message using INFO as log level.
func Info(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

// Debug logs a message using DEBUG as log level.
func Debug(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

// SetLevel for default logger
func SetLevel(lvl string) error {
	level, err := logrus.ParseLevel(lvl)
	if err != nil {
		return err
	}
	logrus.SetLevel(level)
	return nil
}

// Test вызывает функцию, передавая ей буфер-лог в качестве аргумента. Пример использования:
// 	logging.Test(func(log *bytes.Buffer) {
// 		logrus.Into("hello world")
// 		assert.Contains(log.String(), "hello world")
// 	})
func Test(callable func(*bytes.Buffer)) {
	buf := &bytes.Buffer{}
	logrus.SetOutput(buf)

	callable(buf)

	var loggerOut io.Writer
	if std.fd != nil {
		loggerOut = std.fd
	} else {
		loggerOut = os.Stderr
	}

	logrus.SetOutput(loggerOut)
}

// TestWithLevel тоже самое что Test, но можно передать кастомный уровень логирования.
func TestWithLevel(level string, callable func(*bytes.Buffer)) {
	originalLevel := logrus.GetLevel()
	defer logrus.SetLevel(originalLevel)
	SetLevel(level)

	Test(callable)
}

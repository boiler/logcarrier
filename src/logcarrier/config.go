package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

// Config structure
type Config struct {
	Listen      string        `toml:"listen"`
	ListenDebug string        `toml:"listen_debug"`
	WaitTimeout time.Duration `toml:"wait_timeout"`
	Key         string        `toml:"key"`
	DestDir     string        `toml:"destdir"`
	DestDirMode os.FileMode   `toml:"destdir_mode"`
	LogFile     string        `toml:"logfile"`

	Compression struct {
		Method CompressionMethod `toml:"method"`
		Level  uint              `toml:"level"`
	} `toml:"compression"`

	Buffers struct {
		Input   Size `toml:"input"`
		Framing Size `toml:"framing"`
		ZSTDict Size `toml:"zstdict"`

		Connections int `toml:"connections"`
		Dumps       int `toml:"dumps"`
		Logrotates  int `toml:"logrotates"`
	} `toml:"buffers"`

	Workers struct {
		Router     int `toml:"route"`
		Dumper     int `toml:"dumper"`
		Logrotater int `toml:"logrotater"`

		FlusherSleep time.Duration `toml:"flusher_sleep"`
	} `toml:"workers"`
}

// sensible defaults
func initConfig(config *Config) {
	config.Listen = "0.0.0.0:1466"
	config.ListenDebug = ""
	config.WaitTimeout = 60 * time.Second
	config.Key = "key"
	config.DestDir = "./logs"
	config.DestDirMode = 0755
	config.LogFile = ""

	config.Compression.Method = Raw
	config.Compression.Level = 0

	config.Buffers.Input = 128 * 1024
	config.Buffers.Framing = 256 * 1024
	config.Buffers.ZSTDict = 128 * 1024
	config.Buffers.Connections = 1024
	config.Buffers.Dumps = 512
	config.Buffers.Logrotates = 512

	config.Workers.Router = 1024
	config.Workers.Dumper = 24
	config.Workers.Logrotater = 48
	config.Workers.FlusherSleep = time.Second * 30
}

// LoadConfig loads config from given file
func LoadConfig(filePath string) (res Config) {
	var err error
	initConfig(&res)
	defer func() {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read configuration file `\033[1m%s\033[0m`: \033[31m%s\033[0m\n", filePath, err)
			os.Exit(1)
		}
	}()
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}
	if err = toml.Unmarshal(data, &res); err != nil {
		return
	}
	return
}

package config

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

var configFile string
var printDefaultConfig bool

func init() {
	flag.StringVar(&configFile, "config", "", "Filename of config")
	flag.BoolVar(&printDefaultConfig, "config-print-default", false, "Print default config")
}

// PrintConfig печатает конфиг cfg
func PrintConfig(cfg interface{}) error {
	buf := new(bytes.Buffer)

	encoder := toml.NewEncoder(buf)
	encoder.Indent = ""

	if err := encoder.Encode(cfg); err != nil {
		return err
	}

	fmt.Print(buf.String())
	return nil
}

// Parse парсит конфиг из файла, заданного флагом config при запуске приложения.
// Если был указан флаг config-print-default=true, то печатает дефолтный конфиг и завершает выполнение приложения
func Parse(cfg interface{}) error {
	if printDefaultConfig {
		if err := PrintConfig(cfg); err != nil {
			return err
		}
		os.Exit(0)
	}

	if configFile != "" {
		if _, err := toml.DecodeFile(configFile, cfg); err != nil {
			return err
		}
	}

	return nil
}

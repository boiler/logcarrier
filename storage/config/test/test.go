package main

import (
	"flag"
	"log"
	"time"

	"github.com/k0kubun/pp"
	"gitlab.corp.mail.ru/rb/go/config"
)

type mysqlConfig struct {
	Host     string           `toml:"host"`
	Database string           `toml:"database"`
	User     string           `toml:"user"`
	Password string           `toml:"password"`
	Timeout  *config.Duration `toml:"timeout"`
}

type logConfig struct {
	Logfile     string `toml:"logfile"`
	RotateOnHUP bool   `toml:"rotate-on-hup"`
	MaxSize     int    `toml:"max-size"`    // MB
	MaxBackups  int    `toml:"max-backups"` // Count
	MaxAge      int    `toml:"max-age"`     // days
	Localtime   bool   `toml:"localtime"`
}

type sampleConfig struct {
	Logging *logConfig   `toml:"logging"`
	Mysql   *mysqlConfig `toml:"mysql"`
}

func newMysqlConfig() *mysqlConfig {
	config := &mysqlConfig{
		Host:     "localhost",
		Database: "reklama",
		User:     "mpop",
		Password: "",
		Timeout: &config.Duration{
			Duration: 5 * time.Second,
		},
	}
	return config
}

func newLogConfig() *logConfig {
	config := &logConfig{
		Logfile:     "/var/log/rbdatad.log",
		RotateOnHUP: true,
		MaxSize:     500,
		MaxBackups:  0,
		MaxAge:      0,
		Localtime:   true,
	}
	return config
}

func newSampleConfig() *sampleConfig {
	return &sampleConfig{
		Logging: newLogConfig(),
		Mysql:   newMysqlConfig(),
	}
}

func main() {
	flag.Parse()

	cfg := newSampleConfig()
	if err := config.Parse(cfg); err != nil {
		log.Fatal(err.Error())
	}

	pp.Println(cfg)
}

package config

import (
	"log"
	"strings"
	"testing"
	"time"

	"bytes"

	"github.com/BurntSushi/toml"
)

type sampleConfig struct {
	Timeout     *Duration `toml:"timeout"`
	ValueInt    int       `toml:"value-int"`
	ValueString string    `toml:"value-string"`
}

func newSampleConfig() *sampleConfig {
	config := &sampleConfig{}
	config.ValueInt = 42
	config.Timeout = &Duration{Duration: 5 * time.Second}
	config.ValueString = "initial string value"
	return config
}

func TestParse(t *testing.T) {
	config := newSampleConfig()

	if _, err := toml.DecodeFile("sample.toml", config); err != nil {
		t.Fatal(err.Error())
	}

	if config.ValueString != "hello world" {
		t.Fatal("Wrong ValueString value")
	}

	if config.ValueInt != 42 {
		t.Fatal("Wrong ValueInt value")
	}
}

func TestPrintDefault(t *testing.T) {
	buf := new(bytes.Buffer)
	config := newSampleConfig()

	if err := toml.NewEncoder(buf).Encode(config); err != nil {
		log.Fatal(err)
	}
	if !strings.Contains(buf.String(), `timeout = "5s"`) {
		t.Fatal("Encoded timeout not found")
	}

}

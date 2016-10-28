package config

import (
	"time"

	"github.com/BurntSushi/toml"
)

// Duration обертка над time.Duration. Для указания в конфиге в формате "1m30s"
type Duration struct {
	time.Duration
}

var _ toml.TextMarshaler = &Duration{}

// UnmarshalText парсер для TOML-а
func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

// MarshalText кодирует значение в формат "1m30s" для TOML-а
func (d *Duration) MarshalText() ([]byte, error) {
	return []byte(d.Duration.String()), nil
}

// Value возвращает значение типа time.Duration
func (d *Duration) Value() time.Duration {
	return d.Duration
}

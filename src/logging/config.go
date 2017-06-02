package logging

// Config настройки логирования
type Config struct {
	Logfile string `toml:"logfile"`
	Level   string `toml:"level"` // default:"debug"
}

// NewConfig возвращает инстанс Config
func NewConfig() *Config {
	return &Config{
		Level: "info",
	}
}

package main

import "logging"

func main() {
	cfg := LoadConfig("/home/emacs/Sources/logcarrier/test.toml")

	if len(cfg.LogFile) > 0 {
		loggingConfig := logging.NewConfig()
		loggingConfig.Logfile = cfg.LogFile
		if err := logging.SetConfig(loggingConfig); err != nil {
			panic(err)
		}
	}
}

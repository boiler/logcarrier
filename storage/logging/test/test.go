package main

import (
	"flag"
	"time"

	"github.com/Sirupsen/logrus"
	"gitlab.corp.mail.ru/rb/go/logging"
)

func main() {
	logfile := flag.String("log", "", "Logfile")

	flag.Parse()

	logging.SetFile(*logfile)

	logrus.WithFields(logrus.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")

	logging.Debug("debug message")
	logging.Info("info message")
	logging.Warning("warning message")
	logging.Error("error message")
	logging.Critical("critical message")

	for {
		time.Sleep(time.Second)

		logging.Error("message text")
	}

}

package periodic

import (
	"logging"
	"time"
)

// Hourly return a channel which is getting an elemnt once an hour
func Hourly() chan int {
	res := make(chan int, 1024)
	go func() {
		for {
			n := time.Now()
			t := n.Add(time.Hour)
			t = t.Truncate(time.Hour)
			logging.Info("LOGROTATER: will spawn periodic log rotation in %s", t.Sub(n))
			time.Sleep(t.Add(time.Second / 2).Sub(n))
			res <- 0
		}
	}()
	return res
}

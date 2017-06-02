package periodic

import (
	cron "gopkg.in/robfig/cron.v2"
)

// // Hourly return a channel which is getting an element once an hour
// func Hourly() chan int {
// 	res := make(chan int, 1024)
// 	go func() {
// 		for {
// 			n := time.Now()
// 			t := n.Add(time.Hour)
// 			t = t.Truncate(time.Hour)
// 			logging.Info("LOGROTATER: will spawn periodic log rotation in %s", t.Sub(n))
// 			time.Sleep(t.Add(time.Second / 2).Sub(n))
// 			res <- 0
// 		}
// 	}()
// 	return res
// }

// Schedule returns a channel which is getting elements by schedule
func Schedule(sched string) chan int {
	res := make(chan int, 1024)
	c := cron.New()
	c.Start()
	_, err := c.AddFunc(sched, func() {
		res <- 0
	})
	if err != nil {
		panic(err)
	}
	return res
}

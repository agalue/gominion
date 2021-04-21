package tools

import (
	"time"

	"github.com/go-ping/ping"
)

// TODO add logic for handling retries
func Ping(addr string, timeout time.Duration) (time.Duration, error) {
	pinger, err := ping.NewPinger(addr)
	if err != nil {
		return 0, err
	}
	pinger.Count = 1
	pinger.Timeout = timeout
	pinger.SetPrivileged(true)
	if err := pinger.Run(); err != nil {
		return 0, err
	}
	stats := pinger.Statistics()
	return stats.AvgRtt, err
}

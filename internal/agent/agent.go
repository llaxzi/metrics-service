package agent

import (
	"fmt"
	"sync"
	"time"
)

type Agent interface {
	Work()
}

type agent struct {
	pollInterval     int
	reportInterval   int
	pollCount        uint64
	metricsCollector MetricsCollector
	sender           Sender
	metrics          map[string]interface{}
	mu               sync.Mutex
}

func NewAgent(pollInterval int, reportInterval int, metricsCollector MetricsCollector, sender Sender) Agent {
	return &agent{pollInterval, reportInterval, 0, metricsCollector, sender, make(map[string]interface{}), sync.Mutex{}}
}

func (a *agent) Work() {
	pollTicker := time.NewTicker(time.Second * time.Duration(a.pollInterval))

	reportTicker := time.NewTicker(time.Second * time.Duration(a.reportInterval))

	for {
		select {
		case <-pollTicker.C:
			a.mu.Lock()
			a.metrics = a.metricsCollector.Collect()
			a.pollCount++
			a.metrics["PollCount"] = a.pollCount
			fmt.Printf("Collected metrics, pollCount= %d\n", a.pollCount)
			a.mu.Unlock()
		case <-reportTicker.C:
			a.mu.Lock()
			a.sender.Send(a.metrics)
			fmt.Printf("Send metrics \n")
			a.mu.Unlock()
		}
	}
}

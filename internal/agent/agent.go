package agent

import (
	"fmt"
	"metrics-service/internal/agent/collector"
	sender2 "metrics-service/internal/agent/sender"
	"sync"
	"time"
)

type Agent interface {
	Work()
}

type agent struct {
	metricsCollector collector.MetricsCollector
	sender           sender2.Sender
	pollInterval     int
	reportInterval   int
	pollCount        uint64
	mu               sync.Mutex
	metrics          map[string]interface{}
}

func NewAgent(pollInterval int, reportInterval int, metricsCollector collector.MetricsCollector, sender sender2.Sender) Agent {
	return &agent{metricsCollector, sender, pollInterval, reportInterval, 0, sync.Mutex{}, make(map[string]interface{})}
}

// Work Использует Ticker и select для обработки временных интервалов
// Также лочим данные от data race мьютексом
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

			err := a.sender.Send(a.metrics)
			if err != nil {
				fmt.Println(err)

			} else {
				fmt.Println("Send metrics")
				a.pollCount = 0 // Сбрасываем при успешной отправке, т.к. метрика типа counter
			}

			a.mu.Unlock()
		}
	}
}

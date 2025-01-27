package agent

import (
	"fmt"
	"log"
	"metrics-service/internal/agent/collector"
	sender2 "metrics-service/internal/agent/sender"
	"sync"
	"time"
)

type Agent interface {
	Work(doneCh chan struct{})
}

type agent struct {
	metricsCollector collector.MetricsCollector
	sender           sender2.Sender
	pollInterval     int
	reportInterval   int
	pollCount        int64
	mu               sync.Mutex
	metrics          map[string]interface{}
}

func NewAgent(pollInterval int, reportInterval int, metricsCollector collector.MetricsCollector, sender sender2.Sender) Agent {
	return &agent{metricsCollector, sender, pollInterval, reportInterval, 0, sync.Mutex{}, make(map[string]interface{})}
}

// Work Использует Ticker и select для обработки временных интервалов
// Также лочим данные от data race мьютексом
func (a *agent) Work(doneCh chan struct{}) {

	pollTicker := time.NewTicker(time.Second * time.Duration(a.pollInterval))
	reportTicker := time.NewTicker(time.Second * time.Duration(a.reportInterval))

	/*for {
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

			err := a.sender.SendBatch(a.metrics)
			if err != nil {
				log.Println(err)

			} else {
				fmt.Println("Send metrics")
				a.pollCount = 0 // Сбрасываем при успешной отправке, т.к. метрика типа counter
			}

			a.mu.Unlock()
		}
	}*/

	go func() {
		for {
			select {
			case <-pollTicker.C:
				a.mu.Lock()
				a.metrics = a.metricsCollector.Collect()
				a.pollCount++
				a.metrics["PollCount"] = a.pollCount
				fmt.Printf("Collected metrics, pollCount= %d\n", a.pollCount)
				a.mu.Unlock()
			case <-doneCh:
				pollTicker.Stop()
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-reportTicker.C:
				a.mu.Lock()
				err := a.sender.SendBatch(a.metrics)
				if err != nil {
					log.Println(err)
				} else {
					fmt.Println("Send metrics")
					a.pollCount = 0 // Сбрасываем при успешной отправке, т.к. метрика типа counter
				}
				a.mu.Unlock()
			case <-doneCh:
				reportTicker.Stop()
				return
			}
		}
	}()

}

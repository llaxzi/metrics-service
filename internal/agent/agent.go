package agent

import (
	"fmt"
	"log"
	"sync"
	"time"

	"metrics-service/internal/agent/collector"
	sender2 "metrics-service/internal/agent/sender"
)

type Agent interface {
	Collect(doneCh chan struct{})
	ReportBatch(doneCh chan struct{})
	Report(doneCh chan struct{})
}

type agent struct {
	metricsCollector collector.MetricsCollector
	sender           sender2.Sender
	pollInterval     int
	reportInterval   int
	pollCount        int64
	rateLimit        int
	mu               sync.Mutex
	metrics          map[string]interface{}
}

func NewAgent(pollInterval int, reportInterval int, rateLimit int, metricsCollector collector.MetricsCollector, sender sender2.Sender) Agent {
	return &agent{metricsCollector, sender, pollInterval, reportInterval, 0, rateLimit, sync.Mutex{}, make(map[string]interface{})}
}

// Work Использует Ticker и select для обработки временных интервалов
// Также лочим данные от data race мьютексом

func (a *agent) Collect(doneCh chan struct{}) {
	pollTicker := time.NewTicker(time.Second * time.Duration(a.pollInterval))
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
}

func (a *agent) ReportBatch(doneCh chan struct{}) {

	reportTicker := time.NewTicker(time.Second * time.Duration(a.reportInterval))

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

type Metric struct {
	Name  string
	Value interface{}
}

func (a *agent) Report(doneCh chan struct{}) {
	reportTicker := time.NewTicker(time.Second * time.Duration(a.reportInterval))

	// Генератор
	metricsCh := make(chan Metric, 50)
	go func() {
		for {
			select {
			case <-reportTicker.C:
				a.mu.Lock()
				for metricName, metricVal := range a.metrics {
					metricsCh <- Metric{metricName, metricVal}
				}
				a.mu.Unlock()
			case <-doneCh:
				reportTicker.Stop()
				close(metricsCh)
				return
			}
		}
	}()

	// Запуск воркеров
	errCh := make(chan error, 50)
	var wg sync.WaitGroup
	wg.Add(a.rateLimit)
	for i := 0; i < a.rateLimit; i++ {
		go func() {
			defer wg.Done()
			a.worker(metricsCh, errCh, doneCh)
		}()
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Получение ошибок
	for err := range errCh {
		log.Println(err)
	}

}

func (a *agent) worker(metrics <-chan Metric, errCh chan<- error, doneCh chan struct{}) {
	for {
		select {
		case metric := <-metrics:
			err := a.sender.SendJSON(metric.Name, metric.Value)
			if err != nil {
				errCh <- err
				continue
			}
			if metric.Name == "PollCount" {
				a.pollCount = 0
			}
		case <-doneCh:
			return
		}
	}
}

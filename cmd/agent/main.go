package main

import (
	"metrics-service/internal/agent"
	"metrics-service/internal/agent/collector"
	"metrics-service/internal/agent/sender"
)

func main() {

	// Создаем интерфейсы
	metricsCollector := collector.NewMetricsCollector()
	metricsSender := sender.NewSender("http://localhost:8080/update")

	// Конфигурация агента
	pollInterval := 2
	reportInterval := 10

	a := agent.NewAgent(pollInterval, reportInterval, metricsCollector, metricsSender)
	a.Work()

}

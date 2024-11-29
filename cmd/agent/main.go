package main

import (
	"metrics-service/internal/agent"
	"metrics-service/internal/agent/collector"
	"metrics-service/internal/agent/sender"
)

func main() {

	// Парсим флаги
	parseFlags()

	// Создаем интерфейсы
	metricsCollector := collector.NewMetricsCollector()
	metricsSender := sender.NewSender("http://localhost:8080/update")

	// Создаем агент с pollInterval и reportInterval, заданными через flags
	a := agent.NewAgent(pollInterval, reportInterval, metricsCollector, metricsSender)

	// Запускаем агент
	a.Work()

}

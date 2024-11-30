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

	baseURL := "http://" + serverHost + "/update"
	metricsSender := sender.NewSender(baseURL)

	// Создаем агент с pollInterval и reportInterval, заданными через flags
	a := agent.NewAgent(pollInterval, reportInterval, metricsCollector, metricsSender)

	// Запускаем агент
	a.Work()

}

package main

import (
	"metrics-service/internal/agent"
	"metrics-service/internal/agent/collector"
	"metrics-service/internal/agent/sender"
)

func main() {

	// Получаем config (flags или env)
	parseFlags()

	// Создаем интерфейсы
	metricsCollector := collector.NewMetricsCollector()

	baseURL := "http://" + serverHost
	metricsSender := sender.NewSender(baseURL)

	// Создаем агент
	a := agent.NewAgent(pollInterval, reportInterval, metricsCollector, metricsSender)

	// Запускаем агент
	a.Work()

}

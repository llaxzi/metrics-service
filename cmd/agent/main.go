package main

import (
	"fmt"
	"metrics-service/internal/agent"
	"metrics-service/internal/agent/collector"
	"metrics-service/internal/agent/sender"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// Получаем config (flags или env)
	parseFlags()

	// Создаем интерфейсы
	metricsCollector := collector.NewMetricsCollector()

	baseURL := "http://" + serverHost
	metricsSender := sender.NewSender(baseURL, []byte(flagHashKey))

	// Создаем агент
	a := agent.NewAgent(pollInterval, reportInterval, metricsCollector, metricsSender)

	doneCh := make(chan struct{})
	// Перехватываем сигнал Ctrl+C
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// Запускаем агент
	go a.Work(doneCh)

	<-signalCh
	close(doneCh)
	fmt.Println("Agent stopped")

}

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	metricsSender := sender.NewSender(baseURL, []byte(flagHashKey))

	// Создаем агент
	a := agent.NewAgent(pollInterval, reportInterval, flagRateLimit, metricsCollector, metricsSender)

	doneCh := make(chan struct{})
	// Перехватываем сигнал Ctrl+C
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// Запускаем агент
	a.Collect(doneCh)
	if flagReportBatch {
		a.ReportBatch(doneCh)
	} else {
		a.Report(doneCh)
	}

	<-signalCh
	close(doneCh)
	fmt.Println("Agent stopped")

}

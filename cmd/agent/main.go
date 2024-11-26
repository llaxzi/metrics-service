package main

import "metrics-service/internal/agent"

// TODO: pollCount

func main() {
	metricsCollector := agent.NewMetricsCollector()

	sender := agent.NewSender()

	sender.Send(metricsCollector.Collect())

}

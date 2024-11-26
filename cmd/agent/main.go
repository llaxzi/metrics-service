package main

import "metrics-service/internal/agent"

// TODO: pollCount

func main() {

	metricsCollector := agent.NewMetricsCollector()
	sender := agent.NewSender()

	pollInterval := 5
	reportInterval := 10

	a := agent.NewAgent(pollInterval, reportInterval, metricsCollector, sender)
	a.Work()

}

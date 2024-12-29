package service

import "metrics-service/internal/server/storage"

type HtmlService interface {
	GenerateHtml() string
}

type htmlService struct {
	storage storage.MetricsStorage
}

func NewHtmlService(storage storage.MetricsStorage) HtmlService {
	return &htmlService{storage}
}

func (s *htmlService) GenerateHtml() string {
	metrics := s.storage.GetMetrics()

	// Формируем html

	metricsHTML := "<h1>Metrics List</h1><div>"
	if len(metrics) == 0 {
		metricsHTML += "<p>No metrics available</p>"
	} else {
		for _, metric := range metrics {
			metricsHTML += "<p>" + metric[0] + ": " + metric[1] + "</p>" // где metric []string, [0] - metricName, [1] - metricVal
		}
	}
	metricsHTML += "</div>"
	return metricsHTML
}

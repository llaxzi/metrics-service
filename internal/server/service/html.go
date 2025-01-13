package service

import (
	apperrors "metrics-service/internal/server/errors"
	"metrics-service/internal/server/repository"
	"metrics-service/internal/server/storage"
)

type HTMLService interface {
	GenerateHTML() (string, error)
}

type htmlService struct {
	storage       storage.MetricsStorage
	repository    repository.Repository
	useRepository bool
}

func NewHTMLService(storage storage.MetricsStorage, repository repository.Repository, useRepository bool) HTMLService {
	return &htmlService{storage, repository, useRepository}
}

func (s *htmlService) GenerateHTML() (string, error) {
	var metrics [][]string
	if s.useRepository {
		data, err := s.repository.GetSlice()
		if err != nil {
			return "", apperrors.ErrServer
		}
		metrics = data
	} else {
		metrics = s.storage.GetMetrics()
	}

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
	return metricsHTML, nil
}

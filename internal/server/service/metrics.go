package service

import (
	"log"
	apperrors "metrics-service/internal/server/errors"
	"metrics-service/internal/server/models"
	"metrics-service/internal/server/repository"
	"metrics-service/internal/server/retry"
	"metrics-service/internal/server/storage"
	"strconv"
)

type MetricsService interface {
	Update(metricType, metricName, metricValStr string) error
	Get(metricType, metricName string) (string, error)
	UpdateJSON(requestData *models.Metrics) error
	GetJSON(requestData *models.Metrics) error
	Ping() error
	Save() error
	UpdateBatch([]models.Metrics) error
}

type metricsService struct {
	storage         storage.MetricsStorage
	saver           storage.Saver
	repository      repository.Repository
	isStoreInterval bool
	retryer         retry.Retryer
}

func NewMetricsService(storage storage.MetricsStorage, saver interface{ Save() error }, repository repository.Repository, isStoreInterval bool, retryer retry.Retryer) MetricsService {
	return &metricsService{storage, saver, repository, isStoreInterval, retryer}
}

//TODO: логирование ошибок

func (s *metricsService) Update(metricType, metricName, metricValStr string) error {
	// Обновляем значение метрики в зависимости от типа
	switch metricType {
	case "counter":
		metricVal, err := strconv.ParseInt(metricValStr, 10, 64)
		if err != nil {
			return apperrors.ErrWrongMetricValue
		}
		s.storage.SetCounter(metricName, metricVal)

	case "gauge":
		metricVal, err := strconv.ParseFloat(metricValStr, 64)
		if err != nil {
			return apperrors.ErrWrongMetricValue
		}
		s.storage.SetGauge(metricName, metricVal)

	default:
		return apperrors.ErrInvalidMetricType
	}

	return nil
}

func (s *metricsService) Get(metricType, metricName string) (string, error) {

	switch metricType {
	case "counter":
		metricVal, exists := s.storage.GetCounter(metricName)
		if !exists {
			return "", apperrors.ErrMetricNotExist
		}
		return strconv.FormatInt(metricVal, 10), nil
	case "gauge":
		metricVal, exists := s.storage.GetGauge(metricName)

		if !exists {
			return "", apperrors.ErrMetricNotExist
		}
		return strconv.FormatFloat(metricVal, 'f', -1, 64), nil
	default:
		return "", apperrors.ErrInvalidMetricType
	}
}

func (s *metricsService) UpdateJSON(requestData *models.Metrics) error {
	switch requestData.MType {
	case "counter":
		s.storage.SetCounter(requestData.ID, *requestData.Delta)

		actualVal, exists := s.storage.GetCounter(requestData.ID)
		if !exists {
			return apperrors.ErrServer
		}
		*requestData.Delta = actualVal
	case "gauge":
		s.storage.SetGauge(requestData.ID, *requestData.Value)
		actualVal, exists := s.storage.GetGauge(requestData.ID)
		if !exists {
			return apperrors.ErrServer
		}
		*requestData.Value = actualVal
	}
	return nil
}

func (s *metricsService) GetJSON(requestData *models.Metrics) error {
	switch requestData.MType {
	case "counter":
		metricVal, exists := s.storage.GetCounter(requestData.ID)
		if !exists {
			return apperrors.ErrMetricNotExist
		}
		requestData.Delta = &metricVal
	case "gauge":
		metricVal, exists := s.storage.GetGauge(requestData.ID)

		if !exists {
			return apperrors.ErrMetricNotExist
		}
		requestData.Value = &metricVal
	}
	return nil
}

func (s *metricsService) Save() error {
	if s.isStoreInterval {
		return nil
	}
	err := s.saver.Save()
	if err != nil {
		return apperrors.ErrServer
	}
	return nil
}

func (s *metricsService) UpdateBatch(metrics []models.Metrics) error {
	s.storage.SetMetricsJSON(metrics)
	return nil
}

func (s *metricsService) Ping() error {
	if err := s.retryer.Retry(s.repository.Ping); err != nil {
		log.Print(err)
		return apperrors.ErrServer
	}
	return nil
}

package service

import (
	"errors"
	"metrics-service/internal/server/models"
	"metrics-service/internal/server/storage"
	"strconv"
)

type MetricsService interface {
	Update(metricType, metricName, metricValStr string) error
	Get(metricType, metricName string) (string, error)
	UpdateJSON(requestData *models.Metrics) error
	GetJSON(requestData *models.Metrics) error
	SaveDisk() error
}

type metricsService struct {
	storage storage.MetricsStorage
	diskW   storage.DiskWriter
}

func NewMetricsService(storage storage.MetricsStorage, diskW storage.DiskWriter) MetricsService {
	return &metricsService{storage, diskW}
}

func (s *metricsService) Update(metricType, metricName, metricValStr string) error {
	// Обновляем значение метрики в зависимости от типа
	switch metricType {
	case "counter":
		metricVal, err := strconv.ParseInt(metricValStr, 10, 64)
		if err != nil {
			return errors.New("wrong metric value")
		}
		s.storage.SetCounter(metricName, metricVal)

	case "gauge":
		metricVal, err := strconv.ParseFloat(metricValStr, 64)
		if err != nil {
			return errors.New("wrong metric value")
		}
		s.storage.SetGauge(metricName, metricVal)

	default:
		return errors.New("invalid metric type")
	}

	return nil
}

func (s *metricsService) Get(metricType, metricName string) (string, error) {

	switch metricType {
	case "counter":
		metricVal, exists := s.storage.GetCounter(metricName)
		if !exists {
			return "", errors.New("metric doesn't exist")
		}
		return strconv.FormatInt(metricVal, 10), nil
	case "gauge":
		metricVal, exists := s.storage.GetGauge(metricName)

		if !exists {
			return "", errors.New("metric doesn't exist")
		}
		return strconv.FormatFloat(metricVal, 'f', -1, 64), nil
	default:
		return "", errors.New("wrong metric type")
	}
}

func (s *metricsService) UpdateJSON(requestData *models.Metrics) error {
	switch requestData.MType {
	case "counter":
		s.storage.SetCounter(requestData.ID, *requestData.Delta)

		actualVal, exists := s.storage.GetCounter(requestData.ID)
		if !exists {
			return errors.New("server error")
		}
		*requestData.Delta = actualVal
	case "gauge":
		s.storage.SetGauge(requestData.ID, *requestData.Value)
		actualVal, exists := s.storage.GetGauge(requestData.ID)
		if !exists {
			return errors.New("server error")
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
			return errors.New("metric doesn't exist")
		}
		requestData.Delta = &metricVal
	case "gauge":
		metricVal, exists := s.storage.GetGauge(requestData.ID)

		if !exists {
			return errors.New("metric doesn't exist")
		}
		requestData.Value = &metricVal
	}
	return nil
}

func (s *metricsService) SaveDisk() error {
	err := s.diskW.Save()
	if err != nil {
		return errors.New("server error")
	}
	return nil
}

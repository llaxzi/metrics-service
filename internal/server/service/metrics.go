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
	storage       storage.MetricsStorage
	saver         storage.DiskWriter
	repository    repository.Repository
	useRepository bool
	sync          bool
	retryer       retry.Retryer
}

func NewMetricsService(storage storage.MetricsStorage, saver interface{ Save() error }, repository repository.Repository, useRepository bool, isStoreInterval bool, retryer retry.Retryer) MetricsService {
	return &metricsService{storage, saver, repository, useRepository, isStoreInterval, retryer}
}

//TODO: логирование ошибок

func (s *metricsService) Update(metricType, metricName, metricValStr string) error {
	var err error
	if s.useRepository {
		err = s.updateRepo(metricType, metricName, metricValStr)
	} else {
		err = s.updateStorage(metricType, metricName, metricValStr)
	}
	if err != nil {
		return err
	}
	// Сохраняем на диск при синхронном режиме
	err = s.Save()
	if err != nil {
		return err
	}
	return nil
}

func (s *metricsService) Get(metricType, metricName string) (string, error) {
	var data string
	var err error
	if s.useRepository {
		data, err = s.getRepo(metricType, metricName)
	} else {
		data, err = s.getStorage(metricType, metricName)
	}

	if err != nil {
		return "", err
	}
	return data, nil

}

func (s *metricsService) UpdateJSON(requestData *models.Metrics) error {
	var err error
	if s.useRepository {
		err = s.updateJSONRepo(requestData)
	} else {
		err = s.updateJSONStorage(requestData)
	}
	// Сохраняем на диск при синхронном режиме
	err = s.Save()
	if err != nil {
		return err
	}
	return nil
}

func (s *metricsService) GetJSON(requestData *models.Metrics) error {
	var err error
	if s.useRepository {
		err = s.retryer.Retry(func() error {
			return s.getJSONRepo(requestData)
		})
	} else {
		err = s.getJSONStorage(requestData)
	}

	return err
}

func (s *metricsService) Save() error {
	if !s.useRepository && s.sync {
		return nil
	}
	err := s.saver.Save()
	if err != nil {
		return apperrors.ErrServer
	}
	return nil
}

func (s *metricsService) UpdateBatch(metrics []models.Metrics) error {
	var err error
	if s.useRepository {
		err = s.retryer.Retry(func() error {
			return s.repository.Save(metrics)
		})
	} else {
		s.storage.SetMetricsJSON(metrics)
		err = nil
	}

	if err != nil {
		log.Printf("Failed to batch update: %v", err)
		return err
	}
	// Сохраняем на диск при синхронном режиме
	err = s.Save()
	if err != nil {
		log.Printf("Failed to save updates: %v", err)
		return err
	}
	return nil
}

// Repo

func (s *metricsService) Ping() error {
	if err := s.retryer.Retry(s.repository.Ping); err != nil {
		log.Print(err)
		return apperrors.ErrServer
	}
	return nil
}

func (s *metricsService) updateJSONRepo(requestData *models.Metrics) error {
	metrics := []models.Metrics{*requestData}
	err := s.retryer.Retry(func() error {
		return s.repository.Save(metrics)
	})
	if err != nil {
		log.Printf("Failed to update json metric: %v\n", err)
		return apperrors.ErrServer
	}
	return nil
}
func (s *metricsService) getJSONRepo(requestData *models.Metrics) error {
	data, err := s.repository.Get(requestData.MType, requestData.ID)
	if err != nil {
		return err
	}
	requestData.Delta = data.Delta
	requestData.Value = data.Value
	return nil
}

func (s *metricsService) updateRepo(metricType, metricName, metricValStr string) error {
	var delta *int64
	var value *float64
	switch metricType {
	case "counter":
		metricVal, err := strconv.ParseInt(metricValStr, 10, 64)
		if err != nil {
			return apperrors.ErrWrongMetricValue
		}
		delta = &metricVal
		value = nil
	case "gauge":
		metricVal, err := strconv.ParseFloat(metricValStr, 64)
		if err != nil {
			return apperrors.ErrWrongMetricValue
		}
		value = &metricVal

	}
	metrics := []models.Metrics{{metricName, metricType, delta, value}}
	err := s.retryer.Retry(func() error {
		return s.repository.Save(metrics)
	})
	if err != nil {
		return apperrors.ErrServer
	}
	return nil
}
func (s *metricsService) getRepo(metricType, metricName string) (string, error) {
	data, err := s.repository.Get(metricType, metricName)
	if err != nil {
		return "", err
	}
	switch metricType {
	case "counter":
		if data.Delta == nil {
			return "", apperrors.ErrWrongMetricValue
		}
		return strconv.FormatInt(*data.Delta, 10), nil
	case "gauge":
		if data.Value == nil {
			return "", apperrors.ErrWrongMetricValue
		}
		return strconv.FormatFloat(*data.Value, 'f', -1, 64), nil
	default:
		return "", apperrors.ErrInvalidMetricType
	}
}

// Storage
func (s *metricsService) updateJSONStorage(requestData *models.Metrics) error {
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
func (s *metricsService) getJSONStorage(requestData *models.Metrics) error {
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

func (s *metricsService) updateStorage(metricType, metricName, metricValStr string) error {
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
func (s *metricsService) getStorage(metricType, metricName string) (string, error) {
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

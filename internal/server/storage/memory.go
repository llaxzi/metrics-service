package storage

import (
	"context"
	"fmt"
	apperrors "metrics-service/internal/server/errors"
	"metrics-service/internal/server/models"
	"strconv"
	"sync"
)

/*
Хранилище метрик
gauge - метрика текущего состояния системы. Новое значение всегда заменяет старое
counter - метрика-счетчик событий (кол-во запросов и ошибок). Новое значение добавляется к существующему
*/

// Хранилище
type metricsStorage struct {
	muGauge   sync.RWMutex
	muCounter sync.RWMutex
	gauge     map[string]float64
	counter   map[string]int64
	diskW     DiskWriter
}

func (m *metricsStorage) UpdateJSON(ctx context.Context, metric *models.Metrics) error {
	switch metric.MType {
	case "counter":
		m.setCounter(metric.ID, *metric.Delta)
		actualVal, exists := m.getCounter(metric.ID)
		if !exists {
			return apperrors.ErrServer
		}
		*metric.Delta = actualVal
	case "gauge":
		m.setGauge(metric.ID, *metric.Value)
		actualVal, exists := m.getGauge(metric.ID)
		if !exists {
			return apperrors.ErrServer
		}
		*metric.Value = actualVal
	}
	return nil
}

func (m *metricsStorage) Get(ctx context.Context, metricType, metricName string) (string, error) {
	switch metricType {
	case "counter":
		metricVal, exists := m.getCounter(metricName)
		if !exists {
			return "", apperrors.ErrMetricNotExist
		}
		return strconv.FormatInt(metricVal, 10), nil
	case "gauge":
		metricVal, exists := m.getGauge(metricName)

		if !exists {
			return "", apperrors.ErrMetricNotExist
		}
		return strconv.FormatFloat(metricVal, 'f', -1, 64), nil
	default:
		return "", apperrors.ErrInvalidMetricType
	}
}

func (m *metricsStorage) GetJSON(ctx context.Context, metric *models.Metrics) error {
	switch metric.MType {
	case "counter":
		metricVal, exists := m.getCounter(metric.ID)
		if !exists {
			return apperrors.ErrMetricNotExist
		}
		metric.Delta = &metricVal
	case "gauge":
		metricVal, exists := m.getGauge(metric.ID)

		if !exists {
			return apperrors.ErrMetricNotExist
		}
		metric.Value = &metricVal
	}
	return nil
}

func (m *metricsStorage) Update(ctx context.Context, metricType, metricName, metricValStr string) error {
	// Обновляем значение метрики в зависимости от типа
	switch metricType {
	case "counter":
		metricVal, err := strconv.ParseInt(metricValStr, 10, 64)
		if err != nil {
			return apperrors.ErrWrongMetricValue
		}
		m.setCounter(metricName, metricVal)

	case "gauge":
		metricVal, err := strconv.ParseFloat(metricValStr, 64)
		if err != nil {
			return apperrors.ErrWrongMetricValue
		}
		m.setGauge(metricName, metricVal)

	default:
		return apperrors.ErrInvalidMetricType
	}
	return nil
}

func (m *metricsStorage) GetMetrics(ctx context.Context) ([][]string, error) {

	m.muGauge.RLock()
	defer m.muGauge.RUnlock()
	m.muCounter.RLock()
	defer m.muCounter.RUnlock()

	// Используем срез срезов, чтобы хранить одинаковые ключи разных типов
	metrics := make([][]string, 0, len(m.gauge)+len(m.counter)) // len m.gauge и m.counter закрыты мьютексом

	for metricName, metricVal := range m.counter {
		metrics = append(metrics, []string{metricName, strconv.FormatInt(metricVal, 10)})
	}
	for metricName, metricVal := range m.gauge {
		metrics = append(metrics, []string{metricName, strconv.FormatFloat(metricVal, 'f', -1, 64)})
	}
	return metrics, nil
}

func (m *metricsStorage) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			m.setGauge(metric.ID, *metric.Value)
		case "counter":
			m.setCounter(metric.ID, *metric.Delta)
		default:
			return fmt.Errorf("%w, metric: %v", apperrors.ErrInvalidMetricType, metric)
		}
	}
	return nil
}

func (m *metricsStorage) Save() error {
	if m.diskW == nil {
		return nil
	}
	err := m.diskW.Save(m.getMetricsJSON())
	if err != nil {
		// TODO: logging
		return apperrors.ErrServer
	}
	return nil
}

// stub

func (m *metricsStorage) Ping(ctx context.Context) error {
	return apperrors.ErrPingMemory
}

func (m *metricsStorage) Bootstrap(ctx context.Context) error {
	return nil
}

func (m *metricsStorage) Close() error {
	return nil
}

// internal

func (m *metricsStorage) setGauge(key string, value float64) {
	m.muGauge.Lock()
	defer m.muGauge.Unlock()
	m.gauge[key] = value
}

func (m *metricsStorage) getGauge(key string) (float64, bool) {
	m.muGauge.RLock()
	defer m.muGauge.RUnlock()
	val, exists := m.gauge[key]
	return val, exists
}

func (m *metricsStorage) setCounter(key string, value int64) {
	m.muCounter.Lock()
	defer m.muCounter.Unlock()
	m.counter[key] += value
}

func (m *metricsStorage) getCounter(key string) (int64, bool) {
	m.muCounter.RLock()
	defer m.muCounter.RUnlock()
	val, exists := m.counter[key]
	return val, exists
}

func (m *metricsStorage) setMetricsJSON(metrics []models.Metrics) {
	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			m.setGauge(metric.ID, *metric.Value)
		case "counter":
			m.setCounter(metric.ID, *metric.Delta)
		default:
			continue
		}
	}
}

func (m *metricsStorage) getMetricsJSON() []models.Metrics {
	var metrics []models.Metrics

	m.muGauge.RLock()
	defer m.muGauge.RUnlock()
	m.muCounter.RLock()
	defer m.muCounter.RUnlock()

	for name, val := range m.gauge {
		metrics = append(metrics, models.Metrics{ID: name, MType: "gauge", Value: &val})
	}
	for name, val := range m.counter {
		metrics = append(metrics, models.Metrics{ID: name, MType: "counter", Delta: &val})
	}
	return metrics
}

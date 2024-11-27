package storage

/*
Хранилище метрик
gauge - метрика текущего состояния системы. Новое значение всегда заменяет старое
counter - метрика-счетчик событый (кол-во запросов и ошибок). Новое значение добавляется к существующему
*/

// Интерфейс взаимодействия с хранилищем
type MetricsStorage interface {
	SetGauge(key string, value float64)
	GetGauge(key string) (float64, bool)
	SetCounter(key string, value int64)
	GetCounter(key string) (int64, bool)
}

func NewMetricsStorage() MetricsStorage {
	return &metricsStorage{make(map[string]float64), make(map[string]int64)}
}

// Хранилище
type metricsStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func (m *metricsStorage) SetGauge(key string, value float64) {
	m.gauge[key] = value
}

func (m *metricsStorage) GetGauge(key string) (float64, bool) {
	val, exists := m.gauge[key]
	return val, exists
}

func (m *metricsStorage) SetCounter(key string, value int64) {
	m.counter[key] += value
}

func (m *metricsStorage) GetCounter(key string) (int64, bool) {
	val, exists := m.counter[key]
	return val, exists
}

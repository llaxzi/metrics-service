package storage

type GaugeStorage interface {
	Set(key string, value float64)
	Get(key string) (float64, bool)
}

type gaugeStorage struct {
	gauge map[string]float64
}

func NewGaugeStorage() GaugeStorage {
	return &gaugeStorage{make(map[string]float64)}
}

func (s *gaugeStorage) Set(key string, value float64) {
	s.gauge[key] = value
}

func (s *gaugeStorage) Get(key string) (float64, bool) {
	val, exists := s.gauge[key]
	return val, exists
}

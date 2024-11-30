package storage

type CounterStorage interface {
	Set(key string, value int64)
	Get(key string) (int64, bool)
}

type counterStorage struct {
	counter map[string]int64
}

func NewCounterStorage() CounterStorage {
	return &counterStorage{make(map[string]int64)}
}

func (s *counterStorage) Set(key string, value int64) {
	s.counter[key] += value
}

func (s *counterStorage) Get(key string) (int64, bool) {
	val, exists := s.counter[key]
	return val, exists
}

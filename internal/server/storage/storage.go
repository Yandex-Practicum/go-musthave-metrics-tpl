package storage

// Storage структура для хранилища
type Storage struct {
	MemStorage map[string]map[string]interface{}
}

// New создание нового хранилища
func New() *Storage {
	return &Storage{
		MemStorage: make(map[string]map[string]interface{}),
	}
}

// Update обновление метрики
func (s *Storage) Update(metricType, metricName string, metricValue interface{}) error {
	if _, ok := s.MemStorage[metricType]; !ok {
		s.MemStorage[metricType] = make(map[string]interface{})
	}

	s.MemStorage[metricType][metricName] = metricValue
	return nil
}
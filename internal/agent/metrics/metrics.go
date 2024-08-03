package metrics

// Metric структура для хранения метрик
type Metric struct {
	Name  string
	Type  string
	Value interface{}
}

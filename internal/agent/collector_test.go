package agent

import (
	"net/http"
	"sync"
	"testing"
)

func TestNewCollector(t *testing.T) {
	collector := NewCollector()

	if collector == nil {
		t.Fatal("NewCollector() returned nil")
	}

	if collector.gauge == nil {
		t.Error("gauges map is nil")
	}

	if collector.counter == nil {
		t.Error("counters map is nil")
	}
}

func TestUpdateMetrics(t *testing.T) {
	batchSize := 0
	client := http.Client{}
	endpoint := ""
	collector := NewCollector(batchSize, client*http.Client, endpoint)

	// Проверяем начальное состояние
	gauge := collector.GetGauges()
	counter := collector.GetCounters()

	if len(gauge) != 0 {
		t.Errorf("Expected 0 initial gauges, got %d", len(gauge))
	}

	if len(counter) != 0 {
		t.Errorf("Expected 0 initial counters, got %d", len(counter))
	}

	// Обновляем метрики
	collector.UpdateMetrics()

	gauge = collector.GetGauges()
	counter = collector.GetCounters()

	// Проверяем что метрики собрались
	if len(gauge) == 0 {
		t.Error("No gauge metrics collected")
	}

	if len(counter) == 0 {
		t.Error("No counter metrics collected")
	}

	// Проверяем наличие обязательных метрик
	if _, exists := gauge["RandomValue"]; !exists {
		t.Error("RandomValue gauge metric not found")
	}

	if pollCount, exists := counter["PollCount"]; !exists {
		t.Error("PollCount counter metric not found")
	} else if pollCount != 1 {
		t.Errorf("Expected PollCount = 1, got %d", pollCount)
	}

	// Проверяем некоторые runtime метрики
	requiredGauges := []string{
		"Alloc", "Sys", "HeapAlloc", "HeapSys", "NumGC", "GCCPUFraction",
	}

	for _, name := range requiredGauges {
		if _, exists := gauge[name]; !exists {
			t.Errorf("Required gauge metric %s not found", name)
		}
	}
}

func TestPollCountIncrement(t *testing.T) {
	collector := NewCollector()

	// Обновляем метрики несколько раз
	for i := 1; i <= 5; i++ {
		collector.UpdateMetrics()

		counter := collector.GetCounters()
		pollCount := counter["PollCount"]

		if pollCount != int64(i) {
			t.Errorf("After %d updates, expected PollCount = %d, got %d",
				i, i, pollCount)
		}
	}
}

func TestRandomValueChanges(t *testing.T) {
	collector := NewCollector()

	// Собираем несколько значений RandomValue
	const numSamples = 100
	values := make([]float64, numSamples)

	for i := 0; i < numSamples; i++ {
		collector.UpdateMetrics()
		gauges := collector.GetGauges()
		values[i] = gauges["RandomValue"]

		// Проверяем диапазон
		if values[i] < 0 || values[i] >= 1 {
			t.Errorf("RandomValue should be in range [0, 1), got %f", values[i])
		}
	}

	// Проверяем что значения различаются
	// Вероятность того что все 100 значений одинаковы крайне мала
	allSame := true
	firstValue := values[0]
	for i := 1; i < numSamples; i++ {
		if values[i] != firstValue {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("All RandomValue samples are identical, which is statistically very unlikely")
	}

	// Дополнительно проверим что у нас есть разнообразие значений
	// Разделим диапазон [0,1) на 10 интервалов и проверим что заполнены хотя бы 3
	intervals := make([]int, 10)
	for _, value := range values {
		interval := int(value * 10)
		if interval == 10 {
			interval = 9 // Граничный случай для значения 1.0
		}
		intervals[interval]++
	}

	filledIntervals := 0
	for _, count := range intervals {
		if count > 0 {
			filledIntervals++
		}
	}

	if filledIntervals < 3 {
		t.Errorf("Expected at least 3 different intervals to be filled, got %d", filledIntervals)
	}
}

func TestConcurrentAccess(t *testing.T) {
	collector := NewCollector()

	const (
		numWriters      = 10
		numReaders      = 10
		numOpsPerWorker = 100
	)

	var wg sync.WaitGroup

	// Запускаем несколько writer'ов
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOpsPerWorker; j++ {
				collector.UpdateMetrics()
			}
		}()
	}

	// Запускаем несколько reader'ов
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOpsPerWorker; j++ {
				gauges := collector.GetGauges()
				counters := collector.GetCounters()

				// Проверяем базовую целостность данных только если они есть
				// Все обращения к данным должны быть через возвращенные копии
				if len(counters) > 0 {
					if pollCount, exists := counters["PollCount"]; exists && pollCount < 0 {
						t.Errorf("Invalid PollCount: %d", pollCount)
					}
				}

				if len(gauges) > 0 {
					if randomValue, exists := gauges["RandomValue"]; exists {
						if randomValue < 0 || randomValue >= 1 {
							t.Errorf("Invalid RandomValue: %f", randomValue)
						}
					}
				}
			}
		}()
	}

	// Ждем завершения всех операций
	wg.Wait()

	// Проверяем финальное состояние
	finalCounters := collector.GetCounters()
	expectedPollCount := int64(numWriters * numOpsPerWorker)
	if finalCounters["PollCount"] != expectedPollCount {
		t.Errorf("Expected PollCount = %d, got %d", expectedPollCount, finalCounters["PollCount"])
	}

	// Проверяем что данные корректны
	finalGauges := collector.GetGauges()
	if len(finalGauges) == 0 {
		t.Error("No gauge metrics found after concurrent operations")
	}

	if randomValue, exists := finalGauges["RandomValue"]; !exists {
		t.Error("RandomValue not found")
	} else if randomValue < 0 || randomValue >= 1 {
		t.Errorf("Invalid RandomValue: %f", randomValue)
	}
}

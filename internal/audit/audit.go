// Package audit реализует аудит запросов по паттерну «Наблюдатель».
//
// Publisher выступает субъектом (издателем) и рассылает событие аудита
// всем подписанным приёмникам — наблюдателям (Observer). Конкретные
// приёмники: FileSink (запись в файл) и HTTPSink (отправка на удалённый URL).
package audit

import (
	"sync"

	"github.com/rs/zerolog/log"
)

// Event — событие аудита одного запроса с метриками.
type Event struct {
	Ts        int64    `json:"ts"`         // unix timestamp события
	Metrics   []string `json:"metrics"`    // наименования полученных метрик
	IPAddress string   `json:"ip_address"` // IP адрес входящего запроса
}

// Observer — приёмник событий аудита (наблюдатель).
type Observer interface {
	Notify(event Event) error
}

// Publisher — субъект паттерна «Наблюдатель»: хранит набор приёмников
// и рассылает им события аудита.
type Publisher struct {
	mu        sync.RWMutex
	observers []Observer
}

// NewPublisher создаёт издателя с заданным набором приёмников.
func NewPublisher(observers ...Observer) *Publisher {
	return &Publisher{observers: observers}
}

// Subscribe добавляет приёмник аудита.
func (p *Publisher) Subscribe(o Observer) {
	if o == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.observers = append(p.observers, o)
}

// Enabled сообщает, подписан ли хотя бы один приёмник аудита.
// Безопасен для вызова на nil-издателе (аудит отключён).
func (p *Publisher) Enabled() bool {
	if p == nil {
		return false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.observers) > 0
}

// Notify рассылает событие всем приёмникам. Ошибка одного приёмника
// не мешает остальным — она логируется. Безопасен на nil-издателе.
func (p *Publisher) Notify(event Event) {
	if p == nil {
		return
	}
	p.mu.RLock()
	observers := make([]Observer, len(p.observers))
	copy(observers, p.observers)
	p.mu.RUnlock()

	for _, o := range observers {
		if err := o.Notify(event); err != nil {
			log.Error().Err(err).Msg("ошибка отправки события аудита")
		}
	}
}

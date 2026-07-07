// Package audit реализует аудит запросов по паттерну «Наблюдатель».
//
// Publisher выступает субъектом (издателем) и рассылает событие аудита
// всем подписанным приёмникам — наблюдателям (Observer). Конкретные
// приёмники: FileSink (запись в файл) и HTTPSink (отправка на удалённый URL).
//
// Рассылка выполняется асинхронно фиксированным пулом worker-горутин
// через буферизованный канал: это не блокирует хендлер и одновременно
// не даёт количеству отправляющих горутин расти неконтролируемо при
// пиковой нагрузке. Переполнение канала приводит к сбросу события с
// логированием (сеть/файл не поспевают за входящим трафиком).
package audit

import (
	"io"
	"sync"

	"github.com/rs/zerolog/log"
)

// Настройки пула отправки событий.
const (
	// defaultWorkers — число фоновых горутин, обрабатывающих очередь событий.
	defaultWorkers = 4
	// defaultQueueSize — размер буферизованного канала событий.
	defaultQueueSize = 1024
)

// Event — событие аудита одного запроса с метриками.
type Event struct {
	TS        int64    `json:"ts"`         // unix timestamp события
	Metrics   []string `json:"metrics"`    // наименования полученных метрик
	IPAddress string   `json:"ip_address"` // IP адрес входящего запроса
}

// Observer — приёмник событий аудита (наблюдатель).
type Observer interface {
	Notify(event Event) error
}

// Publisher — субъект паттерна «Наблюдатель»: хранит набор приёмников
// и рассылает им события аудита асинхронно через worker pool.
type Publisher struct {
	mu        sync.RWMutex
	observers []Observer

	events   chan Event
	workers  int
	stopOnce sync.Once
	stopped  chan struct{}
	wg       sync.WaitGroup
}

// NewPublisher создаёт издателя с заданным набором приёмников и
// запускает пул фоновых горутин-отправителей.
func NewPublisher(observers ...Observer) *Publisher {
	p := &Publisher{
		observers: observers,
		events:    make(chan Event, defaultQueueSize),
		workers:   defaultWorkers,
		stopped:   make(chan struct{}),
	}
	p.startWorkers()
	return p
}

func (p *Publisher) startWorkers() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *Publisher) worker() {
	defer p.wg.Done()
	for event := range p.events {
		p.dispatch(event)
	}
}

// dispatch синхронно рассылает событие всем текущим приёмникам.
// Ошибка одного приёмника не мешает остальным — она логируется.
func (p *Publisher) dispatch(event Event) {
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

// Notify асинхронно ставит событие в очередь на отправку. Если очередь
// переполнена (приёмники не поспевают), событие отбрасывается с записью
// в лог, чтобы не блокировать вызывающий код (HTTP-хендлер). Безопасен
// на nil-издателе и после Close (в этом случае ничего не делает).
func (p *Publisher) Notify(event Event) {
	if p == nil {
		return
	}
	select {
	case <-p.stopped:
		return
	default:
	}
	select {
	case p.events <- event:
	default:
		log.Warn().Msg("очередь аудита переполнена, событие отброшено")
	}
}

// Close останавливает воркеров, дожидается обработки уже поставленных
// в очередь событий и закрывает приёмники, реализующие io.Closer.
// Безопасен на nil-издателе и при повторном вызове.
func (p *Publisher) Close() error {
	if p == nil {
		return nil
	}
	p.stopOnce.Do(func() {
		close(p.stopped)
		close(p.events)
	})
	p.wg.Wait()

	p.mu.Lock()
	observers := p.observers
	p.observers = nil
	p.mu.Unlock()

	var firstErr error
	for _, o := range observers {
		c, ok := o.(io.Closer)
		if !ok {
			continue
		}
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

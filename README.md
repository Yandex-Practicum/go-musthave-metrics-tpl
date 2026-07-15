# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Сборка с информацией о версии

Бинарники `server` и `agent` выводят при старте в stdout сведения о сборке:

```
Build version: <buildVersion>
Build date: <buildDate>
Build commit: <buildCommit>
```

Значения задаются глобальными переменными `buildVersion`, `buildDate` и
`buildCommit` в пакете `main` каждой команды. Если значение не задано при
компиляции, выводится `N/A`.

Переменные заполняются во время сборки через `-ldflags "-X ..."`:

```
go build -ldflags "-X main.buildVersion=v1.0.0 -X 'main.buildDate=$(date +%Y-%m-%d)' -X main.buildCommit=$(git rev-parse --short HEAD)" -o server ./cmd/server
go build -ldflags "-X main.buildVersion=v1.0.0 -X 'main.buildDate=$(date +%Y-%m-%d)' -X main.buildCommit=$(git rev-parse --short HEAD)" -o agent ./cmd/agent
```

Без флагов бинарник тоже собирается — при запуске все три поля покажут `N/A`.

## Профилирование и оптимизация памяти

Профилирование проводится на **работающем сервере под реалистичной нагрузкой**,
а не только на микробенчмарках: так в профиль попадают и аллокации HTTP-слоя
(gzip, роутер), которые на бенчмарках хранилища не видны.

### Методика

1. В роутер сервера подключён pprof (`r.Mount("/debug", chimw.Profiler())`),
   профиль доступен по `http://<addr>/debug/pprof/heap`.
2. Скрипт `scripts/loadtest.sh`:
   - запускает сервер в режиме синхронной записи на диск (`-i 0 -f <файл>`);
   - наполняет хранилище непустым реалистичным набором метрик (батч gauge+counter);
   - эмулирует нагрузку утилитой [`hey`](https://github.com/rakyll/hey) по
     «тяжёлому» эндпоинту `POST /updates/`;
   - **сразу после нагрузки**, пока память ещё занята, снимает heap-профиль.

```
# базовый профиль (до оптимизаций)
HEY=hey N=300 C=1 M=20 bash scripts/loadtest.sh ./server ./profiles/base.pprof
# профиль после оптимизаций
HEY=hey N=300 C=1 M=20 bash scripts/loadtest.sh ./server ./profiles/result.pprof
```

Профили сохранены в `profiles/base.pprof` и `profiles/result.pprof`.

### Найденные узкие места (live-профиль, `top`/`list`)

Профиль под нагрузкой показал главный источник аллокаций — **создание gzip-writer
на каждый ответ** (`compress/flate.NewWriter` ≈ 89% `alloc_space`), плюс «тяжёлый»
путь сохранения в хранилище:

- `middleware.GzipMiddleware` создавал новый `gzip.Writer` на каждый запрос.
- `MemoryStorage.UpdateBatch` вызывал `UpdateGauge`/`UpdateCounter` по каждой
  метрике, а при синхронизации каждый вызов целиком сериализовал всё хранилище:
  пакет из N метрик → N полных сериализаций в файл.
- `SaveToFile` копировал обе мапы через `GetAllGauges`/`GetAllCounters` и брал
  адрес каждого значения отдельно (мелкие escape-аллокации).
- `ListHandler` использовал `fmt.Sprintf` в цикле.

### Что сделано

- `gzipWriterPool` (`sync.Pool`) переиспользует `gzip.Writer` между запросами
  через `gz.Reset(w)` — почти полностью убирает аллокации `compress/flate`.
- `UpdateBatch` берёт блокировку один раз, пишет в мапы напрямую и синхронизирует
  файл единожды в конце.
- Метод `snapshot()` — один проход под одной `RLock` с предвыделенными
  backing-массивами (`Value`/`Delta` указывают в них без реаллокаций).
- `SaveToFile` использует `snapshot()` и `bufio.Writer`.
- `ListHandler` собирает HTML через `strconv` и `strings.Builder` без `fmt`.

### Результат под нагрузкой (одинаковые параметры `N=300 C=1 M=20`)

Суммарный объём аллокаций (`alloc_space`):

```
base:    375.97MB
result:   89.81MB        (-76%, в ~4.2 раза меньше)
```

### Вывод `pprof -top -diff_base`

```
go tool pprof -top -sample_index=alloc_space -diff_base=profiles/base.pprof profiles/result.pprof

Showing nodes accounting for -283.65MB, 75.45% of 375.97MB total
      flat  flat%   sum%        cum   cum%
 -169.23MB 45.01% 45.01%  -286.23MB 76.13%  compress/flate.NewWriter
  -71.35MB 18.98% 63.99%  -116.99MB 31.12%  compress/flate.(*compressor).init
  -45.14MB 12.01% 76.00%   -45.14MB 12.01%  compress/flate.newDeflateFast (inline)
   25.10MB  6.67% 69.32%    25.10MB  6.67%  bufio.NewWriterSize (inline)
  -19.03MB  5.06% 74.38%     5.07MB  1.35%  storage.SaveToFile
    7.51MB  2.00% 72.39%     7.51MB  2.00%  storage.(*MemoryStorage).snapshot
      -5MB  1.33% 73.72%       -5MB  1.33%  storage.(*MemoryStorage).GetAllCounters
      -3MB   0.8% 74.51%       -3MB   0.8%  storage.(*MemoryStorage).GetAllGauges
```

Значения по ключевым функциям отрицательные, суммарно **−283.65 МБ**. Небольшой
плюс у `snapshot`/`bufio` — это перенос места аллокации из устранённых горячих
путей; общий итог глубоко отрицательный.

### Микробенчмарки (`-benchmem`)

Дополнительно добавлены бенчмарки компонентов (`internal/storage/bench_test.go`,
`internal/handlers/bench_test.go`, `internal/agent/bench_test.go`):

| Бенчмарк                 | B/op (base → result) | allocs/op (base → result) |
|--------------------------|----------------------|---------------------------|
| BenchmarkUpdateBatchSync | 125399 → 11022       | 929 → 31                  |
| BenchmarkSaveToFile      | 48911 → 20057        | 221 → 9                   |

```
go test -run=^$ -bench=. -benchmem -count=1 ./internal/storage
```

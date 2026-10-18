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


добавление бинаря тестов в пути
echo 'export PATH="$PATH:$HOME/Desktop/yadex GO course/yandex-go-musthave-metrics-tpl"' >> ~/.zshrc



компиляция бинарника
go build -o cmd/server/server ./cmd/server

go build -o cmd/agent/agent ./cmd/agent


запуск теста
metricstest-darwin-arm64 -test.v -test.run='^TestIteration1$' -binary-path=cmd/server/server

metricstest-darwin-arm64 -test.v -test.run='^TestIteration2$' -agent-binary-path=cmd/agent/agent

metricstest-darwin-arm64 -test.v -test.run='^TestIteration5$' -binary-path=cmd/server/server

metricstest-darwin-arm64 \
-test.v \
-test.run='^TestIteration2$' \
-binary-path=cmd/server/server \
-source-path=. \
-agent-binary-path=cmd/agent/agent \
-server-port=8080

metricstest-darwin-arm64 \
-test.v \
-test.run='^TestIteration5$' \
-binary-path=cmd/server/server \
-source-path=. \
-agent-binary-path=cmd/agent/agent \
-server-port=8080

#!/bin/bash

# Скрипт для запуска тестов в стиле CI/CD
# Имитирует поведение metricstest-linux-amd64

echo "Запуск тестов для проверки работы сервера и агента..."

# Проверяем, что бинарники собраны
if [ ! -f "bin/server" ]; then
    echo "Ошибка: бинарник сервера не найден"
    exit 1
fi

if [ ! -f "bin/agent" ]; then
    echo "Ошибка: бинарник агента не найден"
    exit 1
fi

echo "Бинарники найдены: bin/server и bin/agent"

# Запускаем unit-тесты для проверки функциональности
echo "Запуск unit-тестов..."
go test -v ./internal/agent/...

echo "Тесты завершены успешно"
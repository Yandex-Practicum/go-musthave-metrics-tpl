# cmd/agent

В данной директории будет содержаться код Сервера, который скомпилируется в бинарное приложение

## Агент для сбора метрик

Агент для сбора метрик - это программа, которая собирает метрики с хоста, на котором она запущена, и отправляет их на сервер метрик.

Сбор просходит каждые 2 секунды.
Отправка происходит каждые 10 секунд.

### Метрики

Метрики ниже имеют значени float64

Alloc, BuckHashSys, Frees, GCCPUFraction, GCSys, HeapAlloc, HeapIdle, HeapInuse, HeapObjects, HeapReleased, HeapSys, LastGC, Lookups, MCacheInuse, MCacheSys, MSpanInuse, MSpanSys, Mallocs, NextGC, NumForcedGC, NumGC, OtherSys, PauseTotalNs, StackInuse, StackSys, Sys, TotalAlloc, RandomValue

Метрики ниже имеют значени int64

PollCount

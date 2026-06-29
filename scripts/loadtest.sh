#!/usr/bin/env bash
# Нагрузочное профилирование сервера метрик.
#
# Запускает сервер в режиме синхронной записи на диск (-i 0), наполняет
# хранилище реалистичным непустым набором метрик, эмулирует нагрузку
# инструментом hey по «тяжёлому» эндпоинту /updates/ и сразу после
# нагрузки (пока память ещё занята) снимает профиль кучи в файл.
#
# Использование:
#   loadtest.sh <server_binary> <out.pprof>
#
# Переменные окружения:
#   HEY   путь к бинарю hey      (по умолчанию ищется в PATH)
#   ADDR  адрес сервера          (по умолчанию 127.0.0.1:8085)
#   N     общее число запросов   (по умолчанию 2000)
#   C     уровень конкуренции    (по умолчанию 1)
#   M     метрик в одном батче   (по умолчанию 100, т.е. 50 gauge + 50 counter)
set -euo pipefail

SERVER_BIN="$1"
OUT="$2"
ADDR="${ADDR:-127.0.0.1:8085}"
HEY="${HEY:-hey}"
N="${N:-2000}"
C="${C:-1}"
M="${M:-100}"
BASEURL="http://$ADDR"

TMP="$(mktemp -d)"
FILE="$TMP/metrics.json"
BODY="$TMP/batch.json"

# Генерируем тело батча: половина gauge, половина counter.
half=$((M / 2))
{
	printf '['
	sep=""
	for i in $(seq 1 "$half"); do
		printf '%s{"id":"Gauge%d","type":"gauge","value":%d.5}' "$sep" "$i" "$i"; sep=","
		printf '%s{"id":"Counter%d","type":"counter","delta":%d}' "$sep" "$i" "$i"
	done
	printf ']'
} >"$BODY"

# Запуск сервера: синхронная запись, без восстановления, тихий лог.
"$SERVER_BIN" -a "$ADDR" -i 0 -f "$FILE" -r=false -l error &
SRV_PID=$!
cleanup() { kill "$SRV_PID" 2>/dev/null || true; rm -rf "$TMP"; }
trap cleanup EXIT

# Ждём готовности сервера.
for _ in $(seq 1 40); do
	if curl -sf "$BASEURL/" >/dev/null 2>&1; then break; fi
	sleep 0.25
done

# Прогрев — наполняем хранилище непустым набором.
curl -s -X POST -H 'Content-Type: application/json' \
	--data-binary @"$BODY" "$BASEURL/updates/" >/dev/null

# Нагрузка по тяжёлому пути /updates/.
"$HEY" -n "$N" -c "$C" -t 60 -m POST -T 'application/json' -D "$BODY" "$BASEURL/updates/" \
	| sed -n '1,18p'

# Снимаем профиль кучи сразу после нагрузки — память ещё занята.
curl -s "$BASEURL/debug/pprof/heap?gc=1" -o "$OUT"
echo "профиль сохранён: $OUT"

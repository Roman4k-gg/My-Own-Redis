# My-Own-Redis

Redis-совместимый in-memory сервер на Go. Поддерживает RESP2, TTL, AOF и
команды PING, ECHO, SET, GET, DEL, EXISTS, INCR.

Проект писался, чтобы разобраться, как Redis устроен внутри: TCP, RESP2,
TTL, AOF и конкурентный доступ.

[![CI](https://github.com/Roman4k-gg/My-Own-Redis/actions/workflows/ci.yml/badge.svg)](https://github.com/Roman4k-gg/My-Own-Redis/actions/workflows/ci.yml)

## Запуск

```bash
git clone https://github.com/Roman4k-gg/My-Own-Redis
cd My-Own-Redis
go run ./cmd/server
```

Подключение через стандартный `redis-cli`:

```bash
redis-cli -p 6379
127.0.0.1:6379> PING
PONG
```

## Как устроено

- `cmd/server` запускает TCP-сервер и graceful shutdown.
- `resp` читает и пишет RESP2.
- `storage` хранит ключи, TTL и запускает фоновый GC.
- `handler` разбирает команды.
- `aof` пишет журнал и восстанавливает состояние после перезапуска.

GC работает в две фазы. Сначала под RLock собирает истёкшие ключи, потом под
WLock удаляет их. GET и SET не блокируются на время сканирования.

## Команды

| Команда | Описание |
|---|---|
| `PING [message]` | Проверка соединения |
| `ECHO message` | Возврат сообщения |
| `SET key value [EX seconds]` | Запись с TTL |
| `GET key` | Чтение |
| `DEL key [key ...]` | Удаление |
| `EXISTS key [key ...]` | Поверка наличия |
| `INCR key` | Атомарный инкремент |

## Тесты

```bash
go test ./...
go test -race ./...
go test -bench=. ./storage
```

Есть конкурентный тест на 200 горутин по 500 операций. CI запускает
`golangci-lint` и `go test -race`. Результаты бенчмарков зависят от CPU и
версии Go, поэтому таблицу лучше собирать на своей машине.

## Ограничения

- Только RESP2, без RESP3.
- Нет репликации и кластера.
- AOF не сжимается.
- Набор команд уже, чем у Redis.

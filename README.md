# Link Storage Service

REST-сервис для хранения и управления короткими ссылками.

## Возможности

- Создание короткой ссылки
- Получение оригинального URL по короткому коду
- Счётчик переходов
- Список всех ссылок с пагинацией
- Удаление ссылки
- Статистика по ссылке
- In-memory кэш с TTL для ускорения чтения
- Конкурентно-безопасная работа

## Стек

- **Go 1.22** (стандартная библиотека + `lib/pq`)
- **PostgreSQL 16** — хранение данных
- **In-memory cache** — кэширование с TTL и фоновой очисткой
- **Docker / docker-compose** — запуск окружения

## Быстрый старт (Docker)

```bash
git clone <repo-url>
cd link-storage-service
docker-compose up --build
```

Сервис будет доступен на `http://localhost:8080`.

## Запуск локально

### Требования

- Go 1.22+
- PostgreSQL (или через Docker)

### 1. Поднять базу данных

```bash
docker run -d \
  --name links-db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=links \
  -p 5432:5432 \
  postgres:16-alpine
```

### 2. Настроить окружение

```bash
cp .env.example .env
# отредактировать .env при необходимости
```

### 3. Запустить сервис

```bash
export $(cat .env | xargs)
go run ./cmd/api
```

## Переменные окружения

| Переменная              | По умолчанию                                              | Описание                        |
|-------------------------|-----------------------------------------------------------|---------------------------------|
| `HTTP_ADDR`             | `:8080`                                                   | Адрес и порт сервера            |
| `DATABASE_URL`          | `postgres://postgres:postgres@localhost:5432/links?sslmode=disable` | Строка подключения к PostgreSQL |
| `CACHE_TTL`             | `60s`                                                     | Время жизни записи в кэше       |
| `CACHE_SWEEP_INTERVAL`  | `2m`                                                      | Интервал очистки устаревших записей кэша |

## API

### Создать короткую ссылку

```
POST /links
```

```json
// Request
{ "url": "https://example.com/some/very/long/url" }

// Response 201
{ "short_code": "aB3kR9" }
```

### Получить ссылку

```
GET /links/{short_code}
```

```json
// Response 200
{ "url": "https://example.com/some/very/long/url", "visits": 15 }
```

Каждый запрос увеличивает счётчик `visits`.

### Список ссылок

```
GET /links?limit=10&offset=0
```

```json
// Response 200
[
  {
    "id": 1,
    "short_code": "aB3kR9",
    "original_url": "https://example.com",
    "created_at": "2024-01-01T12:00:00Z",
    "visits": 15
  }
]
```

### Удалить ссылку

```
DELETE /links/{short_code}
```

```
Response 204 No Content
```

### Статистика

```
GET /links/{short_code}/stats
```

```json
// Response 200
{
  "short_code": "aB3kR9",
  "url": "https://example.com",
  "visits": 15,
  "created_at": "2024-01-01T12:00:00Z"
}
```

## Коды ошибок

| Код | Описание              |
|-----|-----------------------|
| 400 | Некорректный запрос   |
| 404 | Ссылка не найдена     |
| 500 | Внутренняя ошибка     |

## Архитектура

```
cmd/api/          — точка входа, инициализация зависимостей
internal/
  config/         — конфигурация через переменные окружения
  domain/         — модель данных и ошибки
  repository/
    postgres/     — реализация хранилища на PostgreSQL
    cache/        — in-memory кэш с TTL
  service/        — бизнес-логика
  handler/        — HTTP-обработчики
```

# Zal MVP API

Go REST API для сервиса бронирования мест в играх.

## Стек технологий

- **Language:** Go
- **Router:** chi/last version / на ваше усмотрение
- **Database:** PostgreSQL 18 + pgx/v5 (connection pool)
- **Migrations:** golang-migrate/last version
- **Code generation:** sqlc (type-safe SQL)
- **Caching & Locking:** Redis 8.6
- **Observability:** chi middleware (logging, recovery, timeout) 

## Архитектурные особенности

### Защита от овербукинга
( НЕ ГОТОВО, )
- Distributed locks через Redis (`SetNX`)
- Транзакции PostgreSQL с оптимизированными индексами
- Идемпотентные операции

### Защита от двойной оплаты
( НЕ ГОТОВО, )
- `UNIQUE` constraint на `participant_id` в `payments`
- `UNIQUE` constraint на `idempotency_key`
- Status machine: `RESERVED` → `PAID`
- Redis distributed lock на duration платежа

### Генерация кода из SQL

```bash
cd apps/api
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.25.0 generate -f sqlc.yaml
```

Эта команда генерирует type-safe Go код из `queries/*.sql` в `internal/database/sqlc/`.

Важно: при `docker compose up --build` генерация `sqlc` также выполняется автоматически внутри стадии сборки API-образа. Это защищает от ситуации, когда SQL изменили, а generated-код забыли обновить.

### Запуск через Docker Compose

```bash
# Из корня проекта
cd ../..
docker compose up --build

# Логи API
docker compose logs -f api

# Остановка
docker compose down
```

## API Endpoints (НЕ ГОТОВО)

### Public
- `GET /health` - Health check (PostgreSQL + Redis)
- `GET /slots` - Список слотов с фильтрами
- `GET /slots/{id}` - Информация о слоте

### Требуют X-Demo-Token
- `POST /slots/{id}/join` - Присоединиться к слоту
- `POST /slots/{id}/pay` - Оплатить участие
- `GET /me/participations` - Мои участия

### Требуют X-Admin-Token
- `POST /admin/slots` - Создать новый слот

## Миграции

Миграции запускаются автоматически при старте приложения.

## Проверка после запуска

```bash
# 1. Health check
curl http://localhost:8080/health
# Ожидается: {"status":"ok","version":"0.1.0"}

# 2. Подключиться к PostgreSQL
docker compose exec db psql -U postgres -d zal

# Внутри psql:
\dt                    # Список таблиц
\d users               # Структура таблицы users
SELECT COUNT(*) FROM users;  # Должно быть 3 (seed-данные)

# 3. Подключиться к Redis
docker compose exec redis redis-cli
# Внутри redis-cli:
PING                   # Должно вернуть PONG
```

## Переменные окружения

- `DATABASE_URL` - PostgreSQL connection string (обязательно)
- `REDIS_URL` - Redis connection string (обязательно)
- `PORT` - Порт HTTP сервера (по умолчанию: 8080)
- `ENV` - Окружение: development/production (по умолчанию: development)
- `ADMIN_TOKEN` - Токен для админских операций
- `DEMO_USER_TOKEN` - Демо-токен для тестирования

## License

См. корневой README.md

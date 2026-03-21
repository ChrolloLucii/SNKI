# Zal MVP — Слоты для командных видов спорта

> Сервис для игроков без команды, который помогает быстро найти и занять место в игре в ближайшем спортзале с прозрачными правилами и предсказуемой стоимостью.

---

## Содержание

1. [Обзор проекта](#1-обзор-проекта)
2. [Объем MVP](#2-объем-mvp)
3. [Архитектура](#3-архитектура)
4. [Как запустить локально (Docker Compose)](#4-как-запустить-локально-docker-compose)
5. [Как запускать миграции](#5-как-запускать-миграции)
6. [Как заполнить демо-данные](#6-как-заполнить-демо-данные)
7. [Сводка API эндпоинтов](#7-сводка-api-эндпоинтов)
8. [Процесс внесения изменений](#8-процесс-внесения-изменений)
9. [Процесс работы с issue и доской](#9-процесс-работы-с-issue-и-доской)

---

## 1. Обзор проекта

**Zal** помогает игрокам без команды находить свободные игровые слоты в местных залах, присоединяться к ним и оплачивать участие — прямо с телефона.

Какие проблемы решает сервис:
- Сложно найти любительские игры в ближайших залах
- Непрозрачная стоимость и правила отмены
- Неявки и отмены в последний момент срывают игры

Пользователь открывает PWA на телефоне, выбирает вид спорта, район и время, затем просматривает доступные слоты, читает правила и бронирует место в один тап.

---

## 2. Объем MVP

### ✅ Обязательно
- Каталог слотов с фильтрами (вид спорта, район, диапазон дат)
- Страница деталей слота («карточка слота») с правилами, ценой и политикой отмены
- Присоединение/бронирование места (с защитой от гонок и овербукинга)
- Имитация оплаты (переводит статус участия в `PAID`)
- Страница «Мои участия»
- Admin API для создания слотов + наполнение демо-данными

### 🟡 Желательно
- Базовые уведомления внутри приложения
- Admin веб-страница для создания слотов

### ❌ Вне рамок MVP
- Реальный платежный провайдер (Stripe и др.)
- Push-уведомления
- Микросервисы / Kafka / gRPC
- Нативные iOS / Android приложения (только PWA)

---

## 3. Архитектура

Диаграммы переведены в Mermaid и вынесены в [docs/mermaid.md](docs/mermaid.md).

Подробности по архитектуре: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

---

## 4. Как запустить локально (Docker Compose)

**Требования:** [Docker Desktop](https://www.docker.com/products/docker-desktop/) (или Docker + Docker Compose v2).

```bash
# 1. Клонируйте репозиторий
git clone https://github.com/<your-org>/zal-mvp.git
cd zal-mvp

# 2. Скопируйте переменные окружения
cp .env.example .env

# 3. Запустите всё (DB + API + Web)
docker compose up --build
```

Сервисы будут доступны по адресам:

| Сервис | URL |
|---------|-----|
| Web PWA | http://localhost:5173 |
| API     | http://localhost:8080 |
| DB      | postgresql://postgres:postgres@localhost:5432/zal |

Остановить: `docker compose down`  
Сбросить БД: `docker compose down -v`

---

## 5. Как запускать миграции

Миграции запускаются автоматически при старте API через `golang-migrate`.

Для ручного запуска:

```bash
# Внутри запущенного контейнера api
docker compose exec api ./migrate -path /app/migrations -database "$DATABASE_URL" up

# Или через локально установленный migrate CLI
migrate -path apps/api/migrations -database "postgresql://postgres:postgres@localhost:5432/zal?sslmode=disable" up
```

Файлы миграций находятся в `apps/api/migrations/`.

---

## 6. Как заполнить демо-данные

```bash
# Заполнить через admin API (нужен ADMIN_TOKEN из .env)
curl -X POST http://localhost:8080/admin/slots \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: secret" \
  -d '{
    "sport": "football",
    "district": "Центральный",
    "venue_name": "Стадион Динамо",
    "address": "ул. Ленина 1",
    "starts_at": "2026-04-01T18:00:00Z",
    "duration_minutes": 60,
    "capacity": 10,
    "min_players": 6,
    "deadline_at": "2026-04-01T16:00:00Z",
    "expected_price": 500,
    "max_price": 700,
    "rules_text": "Casual game, no tackles."
  }'
```

Скрипт наполнения (для CI / локального демо) находится в `apps/api/cmd/seed/main.go`.

---

## 7. Сводка API эндпоинтов

| Метод | Path | Auth | Описание |
|--------|------|------|-------------|
| GET | `/health` | — | Проверка доступности |
| GET | `/slots` | — | Список слотов (`?sport=&district=&date_from=&date_to=`) |
| GET | `/slots/{slotId}` | — | Детали слота |
| POST | `/slots/{slotId}/join` | User token | Забронировать место (409 если мест нет) |
| POST | `/slots/{slotId}/pay` | User token | Имитация оплаты → PAID |
| GET | `/me/participations` | User token | Мои бронирования |
| POST | `/admin/slots` | Admin token | Создать слот |

Полная спецификация API: [docs/API.md](docs/API.md)

---

## 8. Процесс внесения изменений

Мы используем **trunk-based development** с короткоживущими ветками.

```
main  ←── feature/<ticket>-short-desc  (PR + review)
      ←── fix/<ticket>-short-desc
      ←── tech/<ticket>-short-desc
```

### Шаги

1. Возьмите issue из колонки **Ready** на project board.
2. Создайте ветку: `git checkout -b feature/42-slot-filters`
3. Внесите изменения, коммитьте чаще с понятными сообщениями:  
   `git commit -m "feat(slots): add district filter"`
4. Запушьте ветку и откройте Pull Request по [шаблону PR](.github/pull_request_template.md).
5. Запросите ревью минимум у одного участника команды.
6. После одобрения объедините в `main` (предпочтительно squash merge).

### Соглашение по commit messages

```
<type>(<scope>): <short description>

Типы: feat, fix, refactor, test, docs, chore
Области: slots, participation, payments, admin, web, db, infra
```

---

## 9. Процесс работы с issue и доской

Используем GitHub Issues + GitHub Project board с такими колонками:

| Колонка | Значение |
|--------|---------|
| **Backlog** | Идеи и будущие задачи — без приоритизации |
| **Ready** | Задачи с понятным объемом, готовые к работе |
| **In Progress** | Кто-то активно выполняет задачу |
| **In Review** | PR открыт и ожидает ревью |
| **Done** | Изменения влиты в main |

### Для нетехнических участников команды

- **Сообщить о баге:** создайте issue → выберите шаблон _Bug Report_.
- **Запросить фичу:** создайте issue → выберите шаблон _Feature Request_.
- **Отследить техническую задачу:** используйте шаблон _Tech Task_.
- Назначьте labels (см. ниже) и переместите карточку в нужную колонку.

### Labels

| Label | Значение |
|-------|---------|
| `type:feature` | Новая фича или пользовательская история |
| `type:bug` | Что-то сломано |
| `type:tech` | Рефакторинг, инфраструктура, инструменты |
| `type:docs` | Документация |
| `priority:must` | Блокер для MVP |
| `priority:should` | Важно, но не блокирует |
| `area:web` | Фронтенд |
| `area:api` | Бэкенд |
| `area:db` | База данных / миграции |
| `good first issue` | Подходит новичкам |

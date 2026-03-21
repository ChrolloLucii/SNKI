# Архитектура

## Обзор

Zal MVP — это монорепозиторий с тремя запускаемыми сервисами, связанными через Docker Compose.

Диаграммы вынесены в Mermaid: [docs/mermaid.md](./mermaid.md)

## Структура директорий

```
zal-mvp/
├── apps/
│   ├── api/          # Go REST API
│   │   ├── cmd/api/  # Точка входа (main.go)
│   │   ├── internal/ # Доменные модули: slots, participation, payments, admin
│   │   └── migrations/
│   └── web/          # React + Vite PWA
│       └── src/
│           ├── routes/     # SlotsList, SlotDetails, MyParticipations
│           └── components/
├── infra/            # Общая инфраструктурная конфигурация (зарезервировано)
├── docs/             # ARCHITECTURE.md, API.md, mermaid.md
├── .github/
│   ├── ISSUE_TEMPLATE/
│   └── pull_request_template.md
├── docker-compose.yml
└── .env.example
```

## Ключевые архитектурные решения

| Вопрос | Решение | Причина |
|---------|----------|--------|
| Auth | Демо-токен (без пароля) | Быстрее всего для MVP, без сложной авторизации |
| Овербукинг | Транзакция БД + `SELECT … FOR UPDATE` + проверка вместимости | Корректно при конкурентной нагрузке |
| Платежи | Имитация (только смена статуса) | Реальные платежи вне рамок MVP |
| Стиль API | REST + JSON | Просто и знакомо всем |
| Миграции | golang-migrate, запуск при старте | В dev не нужны дополнительные инструменты |
| Роутинг фронтенда | react-router v6 | Стандартное и хорошо документированное решение |

## Поток данных: присоединение к слоту

Диаграмма потока данных вынесена в Mermaid: [docs/mermaid.md](./mermaid.md)

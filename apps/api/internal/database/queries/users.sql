-- name: GetUserByToken :one
-- Поиск пользователя по демо-токену (используется для аутентификации в каждом запросе)
SELECT id, token, created_at
FROM users
WHERE token = $1
LIMIT 1;

-- name: GetUserByID :one
-- Получение пользователя по UUID
SELECT id, token, created_at
FROM users
WHERE id = $1
LIMIT 1;

-- name: CreateUser :one
-- Создание нового пользователя (для future feature: регистрация)
INSERT INTO users (token)
VALUES ($1)
RETURNING id, token, created_at;

-- name: ListUsers :many
-- Список всех пользователей (для админки)
SELECT id, token, created_at
FROM users
ORDER BY created_at DESC;

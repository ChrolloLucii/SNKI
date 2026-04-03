-- name: GetParticipant :one
-- Получение участия по ID
SELECT
    id, slot_id, user_id,
    status, reserved_at, paid_at
FROM participants
WHERE id = $1
LIMIT 1;

-- name: GetParticipantBySlotAndUser :one
-- Проверка существует ли участие пользователя в слоте
-- Используется перед INSERT для проверки дублей
SELECT
    id, slot_id, user_id,
    status, reserved_at, paid_at
FROM participants
WHERE slot_id = $1 AND user_id = $2
LIMIT 1;

-- name: CreateParticipant :one
-- КРИТИЧНЫЙ: создание участия (POST /slots/{id}/join)
-- Должно выполняться в транзакции после проверки capacity
INSERT INTO participants (slot_id, user_id, status)
VALUES ($1, $2, COALESCE($3, 'RESERVED'))
RETURNING
    id, slot_id, user_id,
    status, reserved_at, paid_at;

-- name: UpdateParticipantStatus :one
-- Обновление статуса участия (RESERVED -> PAID после оплаты)
UPDATE participants
SET
    status = $2,
    paid_at = CASE WHEN $2 = 'PAID' THEN NOW() ELSE paid_at END
WHERE id = $1
RETURNING
    id, slot_id, user_id,
    status, reserved_at, paid_at;

-- name: UpdateParticipantStatusWithCheck :one
-- КРИТИЧНЫЙ: обновление статуса с проверкой текущего статуса
-- Защита от двойной оплаты на уровне запроса
-- WHERE status = $3 гарантирует, что обновление произойдет только если статус совпадает
UPDATE participants
SET
    status = $2,
    paid_at = CASE WHEN $2 = 'PAID' THEN NOW() ELSE paid_at END
WHERE id = $1 AND status = $3
RETURNING
    id, slot_id, user_id,
    status, reserved_at, paid_at;

-- name: ListUserParticipations :many
-- Список всех участий пользователя (GET /me/participations)
-- Включает информацию о слотах через JOIN
SELECT
    p.id, p.slot_id, p.user_id,
    p.status, p.reserved_at, p.paid_at,
    s.sport, s.district, s.venue_name, s.address,
    s.starts_at, s.deadline_at, s.duration_minutes,
    s.expected_price, s.max_price, s.status as slot_status
FROM participants p
JOIN slots s ON p.slot_id = s.id
WHERE p.user_id = $1
ORDER BY s.starts_at DESC;

-- name: ListSlotParticipants :many
-- Список участников конкретного слота (для админки)
SELECT
    p.id, p.slot_id, p.user_id,
    p.status, p.reserved_at, p.paid_at,
    u.token as user_token
FROM participants p
JOIN users u ON p.user_id = u.id
WHERE p.slot_id = $1
ORDER BY p.reserved_at ASC;

-- name: DeleteParticipant :exec
-- Удаление участия (отмена бронирования)
DELETE FROM participants
WHERE id = $1;

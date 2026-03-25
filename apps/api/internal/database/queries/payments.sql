-- name: GetPayment :one
-- Получение платежа по ID
SELECT
    id, participant_id, idempotency_key,
    amount, currency, provider, status,
    provider_payment_id, provider_metadata,
    created_at, updated_at
FROM payments
WHERE id = $1
LIMIT 1;

-- name: GetPaymentByParticipant :one
-- Получение платежа по ID участия
-- Используется для проверки: уже оплачено или нет
SELECT
    id, participant_id, idempotency_key,
    amount, currency, provider, status,
    provider_payment_id, provider_metadata,
    created_at, updated_at
FROM payments
WHERE participant_id = $1
LIMIT 1;

-- name: GetPaymentByIdempotencyKey :one
-- КРИТИЧНЫЙ: поиск платежа по idempotency key
-- Защита от повторных запросов: если ключ найден, возвращаем cached result
SELECT
    id, participant_id, idempotency_key,
    amount, currency, provider, status,
    provider_payment_id, provider_metadata,
    created_at, updated_at
FROM payments
WHERE idempotency_key = $1
LIMIT 1;

-- name: CreatePayment :one
-- КРИТИЧНЫЙ: создание платежа (POST /slots/{id}/pay)
-- Должно выполняться в транзакции вместе с обновлением participants.status
-- UNIQUE constraints на (participant_id, idempotency_key) защищают от дублей
INSERT INTO payments (
    participant_id,
    idempotency_key,
    amount,
    currency,
    provider,
    status,
    provider_payment_id,
    provider_metadata
) VALUES (
    $1, $2, $3,
    COALESCE($4, 'RUB'),
    COALESCE($5, 'FAKE'),
    COALESCE($6, 'PENDING'),
    $7, $8
)
RETURNING
    id, participant_id, idempotency_key,
    amount, currency, provider, status,
    provider_payment_id, provider_metadata,
    created_at, updated_at;

-- name: UpdatePaymentStatus :one
-- Обновление статуса платежа (PENDING -> PAID/FAILED)
-- Используется при обработке webhooks от провайдера
UPDATE payments
SET
    status = $2,
    provider_payment_id = COALESCE($3, provider_payment_id),
    provider_metadata = COALESCE($4, provider_metadata),
    updated_at = NOW()
WHERE id = $1
RETURNING
    id, participant_id, idempotency_key,
    amount, currency, provider, status,
    provider_payment_id, provider_metadata,
    created_at, updated_at;

-- name: GetPaymentByProviderID :one
-- Поиск платежа по ID провайдера (для webhook обработки)
SELECT
    id, participant_id, idempotency_key,
    amount, currency, provider, status,
    provider_payment_id, provider_metadata,
    created_at, updated_at
FROM payments
WHERE provider_payment_id = $1
LIMIT 1;

-- name: ListPaymentsByStatus :many
-- Список платежей по статусу (для аналитики и reconciliation)
SELECT
    id, participant_id, idempotency_key,
    amount, currency, provider, status,
    provider_payment_id, provider_metadata,
    created_at, updated_at
FROM payments
WHERE status = $1
ORDER BY created_at DESC;

-- name: ListUserPayments :many
-- Список всех платежей пользователя (через participants)
SELECT
    py.id, py.participant_id, py.idempotency_key,
    py.amount, py.currency, py.provider, py.status,
    py.provider_payment_id, py.provider_metadata,
    py.created_at, py.updated_at,
    pt.slot_id, s.venue_name, s.starts_at
FROM payments py
JOIN participants pt ON py.participant_id = pt.id
JOIN slots s ON pt.slot_id = s.id
WHERE pt.user_id = $1
ORDER BY py.created_at DESC;

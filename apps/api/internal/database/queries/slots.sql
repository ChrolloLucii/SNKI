-- name: GetSlot :one
-- Получение одного слота по ID для GET /slots/{id}
SELECT
    id, sport, district, venue_name, address,
    starts_at, deadline_at, duration_minutes,
    capacity, min_players,
    expected_price, max_price,
    rules_text, status,
    created_at, updated_at
FROM slots
WHERE id = $1
LIMIT 1;

-- name: ListSlotsWithFilters :many
-- Список слотов с фильтрацией для GET /slots
-- КРИТИЧНЫЙ: использует idx_slots_filter для производительности
SELECT
    id, sport, district, venue_name, address,
    starts_at, deadline_at, duration_minutes,
    capacity, min_players,
    expected_price, max_price,
    rules_text, status,
    created_at, updated_at
FROM slots
WHERE
    status = 'OPEN'
    AND ($1::varchar IS NULL OR sport = $1)
    AND ($2::varchar IS NULL OR district = $2)
    AND ($3::timestamp IS NULL OR starts_at >= $3)
    AND ($4::timestamp IS NULL OR starts_at <= $4)
ORDER BY starts_at ASC;

-- name: CreateSlot :one
-- Создание нового слота (POST /admin/slots)
INSERT INTO slots (
    sport, district, venue_name, address,
    starts_at, deadline_at, duration_minutes,
    capacity, min_players,
    expected_price, max_price,
    rules_text, status
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7,
    $8, $9,
    $10, $11,
    $12, COALESCE($13, 'OPEN')
)
RETURNING
    id, sport, district, venue_name, address,
    starts_at, deadline_at, duration_minutes,
    capacity, min_players,
    expected_price, max_price,
    rules_text, status,
    created_at, updated_at;

-- name: UpdateSlotStatus :one
-- Обновление статуса слота (OPEN -> CANCELLED/COMPLETED)
UPDATE slots
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id, sport, district, venue_name, address,
    starts_at, deadline_at, duration_minutes,
    capacity, min_players,
    expected_price, max_price,
    rules_text, status,
    created_at, updated_at;

-- name: CountSlotParticipants :one
-- КРИТИЧНЫЙ: подсчет участников слота для защиты от овербукинга
-- Используется с FOR UPDATE в транзакции
SELECT COUNT(*)::int
FROM participants
WHERE slot_id = $1
  AND status IN ('RESERVED', 'PAID');

-- name: GetSlotWithParticipantsCount :one
-- Слот с количеством участников (для отображения "5/10 мест занято")
SELECT
    s.id, s.sport, s.district, s.venue_name, s.address,
    s.starts_at, s.deadline_at, s.duration_minutes,
    s.capacity, s.min_players,
    s.expected_price, s.max_price,
    s.rules_text, s.status,
    s.created_at, s.updated_at,
    COUNT(p.id) FILTER (WHERE p.status IN ('RESERVED', 'PAID'))::int as current_participants
FROM slots s
LEFT JOIN participants p ON s.id = p.slot_id
WHERE s.id = $1
GROUP BY s.id;

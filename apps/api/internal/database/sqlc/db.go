package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Slot represents game slot
type Slot struct {
	ID              uuid.UUID
	Sport           string
	District        string
	VenueName       string
	Address         string
	StartsAt        time.Time
	DeadlineAt      time.Time
	DurationMinutes int32
	Capacity        int32
	MinPlayers      int32
	ExpectedPrice   int32
	MaxPrice        int32
	RulesText       sql.NullString
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Participant represents slot participant
type Participant struct {
	ID         uuid.UUID
	SlotID     uuid.UUID
	UserID     uuid.UUID
	Status     string
	ReservedAt time.Time
	PaidAt     sql.NullTime
}

// DB provides database operations
type DB struct {
	pool *pgxpool.Pool
}

// New creates new DB
func New(pool *pgxpool.Pool) *DB {
	return &DB{pool: pool}
}

// GetSlot gets slot by ID
func (d *DB) GetSlot(ctx context.Context, id uuid.UUID) (*Slot, error) {
	query := `
		SELECT
			id, sport, district, venue_name, address,
			starts_at, deadline_at, duration_minutes,
			capacity, min_players,
			expected_price, max_price,
			rules_text, status,
			created_at, updated_at
		FROM slots
		WHERE id = $1
		LIMIT 1
	`

	row := d.pool.QueryRow(ctx, query, id)
	slot := &Slot{}

	err := row.Scan(
		&slot.ID, &slot.Sport, &slot.District, &slot.VenueName, &slot.Address,
		&slot.StartsAt, &slot.DeadlineAt, &slot.DurationMinutes,
		&slot.Capacity, &slot.MinPlayers,
		&slot.ExpectedPrice, &slot.MaxPrice,
		&slot.RulesText, &slot.Status,
		&slot.CreatedAt, &slot.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Слот не найден
		}
		return nil, fmt.Errorf("failed to get slot: %w", err)
	}

	return slot, nil
}

// GetSlotWithParticipantsCount получает слот с количеством участников
func (d *DB) GetSlotWithParticipantsCount(ctx context.Context, id uuid.UUID) (*SlotWithParticipantsCount, error) {
	query := `
		SELECT
			s.id, s.sport, s.district, s.venue_name, s.address,
			s.starts_at, s.deadline_at, s.duration_minutes,
			s.capacity, s.min_players,
			s.expected_price, s.max_price,
			s.rules_text, s.status,
			s.created_at, s.updated_at,
			COUNT(p.id) FILTER (WHERE p.status IN ('RESERVED','PAID'))::int as current_participants
		FROM slots s
		LEFT JOIN participants p ON s.id = p.slot_id
		WHERE s.id = $1
		GROUP BY s.id
	`

	row := d.pool.QueryRow(ctx, query, id)
	slot := &SlotWithParticipantsCount{}

	err := row.Scan(
		&slot.ID, &slot.Sport, &slot.District, &slot.VenueName, &slot.Address,
		&slot.StartsAt, &slot.DeadlineAt, &slot.DurationMinutes,
		&slot.Capacity, &slot.MinPlayers,
		&slot.ExpectedPrice, &slot.MaxPrice,
		&slot.RulesText, &slot.Status,
		&slot.CreatedAt, &slot.UpdatedAt,
		&slot.CurrentParticipants,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get slot with participants count: %w", err)
	}

	return slot, nil
}

// ListSlotsWithFilters получает список слотов с фильтрацией
func (d *DB) ListSlotsWithFilters(ctx context.Context, sport *string, district *string, startsAtFrom *string, startsAtTo *string) ([]Slot, error) {
	query := `
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
			AND ($3::timestamp IS NULL OR starts_at >= to_timestamp($3, 'YYYY-MM-DD HH24:MI:SS'))
			AND ($4::timestamp IS NULL OR starts_at <= to_timestamp($4, 'YYYY-MM-DD HH24:MI:SS'))
		ORDER BY starts_at ASC
	`

	rows, err := d.pool.Query(ctx, query, sport, district, startsAtFrom, startsAtTo)
	if err != nil {
		return nil, fmt.Errorf("failed to list slots: %w", err)
	}
	defer rows.Close()

	slots := []Slot{}
	for rows.Next() {
		slot := Slot{}
		err := rows.Scan(
			&slot.ID, &slot.Sport, &slot.District, &slot.VenueName, &slot.Address,
			&slot.StartsAt, &slot.DeadlineAt, &slot.DurationMinutes,
			&slot.Capacity, &slot.MinPlayers,
			&slot.ExpectedPrice, &slot.MaxPrice,
			&slot.RulesText, &slot.Status,
			&slot.CreatedAt, &slot.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan slot: %w", err)
		}
		slots = append(slots, slot)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return slots, nil
}

// CreateSlot создает новый слот
func (d *DB) CreateSlot(ctx context.Context, arg *CreateSlotParams) (*Slot, error) {
	query := `
		INSERT INTO slots (
			sport, district, venue_name, address,
			starts_at, deadline_at, duration_minutes,
			capacity, min_players,
			expected_price, max_price,
			rules_text, status
		) VALUES (
			$1, $2, $3, $4,
			to_timestamp($5, 'YYYY-MM-DD HH24:MI:SS'), to_timestamp($6, 'YYYY-MM-DD HH24:MI:SS'), $7,
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
			created_at, updated_at
	`

	row := d.pool.QueryRow(ctx, query,
		arg.Sport, arg.District, arg.VenueName, arg.Address,
		arg.StartsAt, arg.DeadlineAt, arg.DurationMinutes,
		arg.Capacity, arg.MinPlayers,
		arg.ExpectedPrice, arg.MaxPrice,
		arg.RulesText, arg.Status,
	)

	slot := &Slot{}
	err := row.Scan(
		&slot.ID, &slot.Sport, &slot.District, &slot.VenueName, &slot.Address,
		&slot.StartsAt, &slot.DeadlineAt, &slot.DurationMinutes,
		&slot.Capacity, &slot.MinPlayers,
		&slot.ExpectedPrice, &slot.MaxPrice,
		&slot.RulesText, &slot.Status,
		&slot.CreatedAt, &slot.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create slot: %w", err)
	}

	return slot, nil
}

// UpdateSlotStatus обновляет статус слота
func (d *DB) UpdateSlotStatus(ctx context.Context, id uuid.UUID, status string) (*Slot, error) {
	query := `
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
			created_at, updated_at
	`

	row := d.pool.QueryRow(ctx, query, id, status)
	slot := &Slot{}

	err := row.Scan(
		&slot.ID, &slot.Sport, &slot.District, &slot.VenueName, &slot.Address,
		&slot.StartsAt, &slot.DeadlineAt, &slot.DurationMinutes,
		&slot.Capacity, &slot.MinPlayers,
		&slot.ExpectedPrice, &slot.MaxPrice,
		&slot.RulesText, &slot.Status,
		&slot.CreatedAt, &slot.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update slot status: %w", err)
	}

	return slot, nil
}

// CountSlotParticipants считает количество участников слота
// метод для защиты от овербукинга
func (d *DB) CountSlotParticipants(ctx context.Context, slotID uuid.UUID) (int32, error) {
	query := `
		SELECT COUNT(*)::int
		FROM participants
		WHERE slot_id = $1
		  AND status IN ('RESERVED', 'PAID')
	`

	var count int32
	err := d.pool.QueryRow(ctx, query, slotID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count participants: %w", err)
	}

	return count, nil
}

// GetParticipant получает участие по ID
func (d *DB) GetParticipant(ctx context.Context, id uuid.UUID) (*Participant, error) {
	query := `
		SELECT
			id, slot_id, user_id,
			status, reserved_at, paid_at
		FROM participants
		WHERE id = $1
		LIMIT 1
	`

	row := d.pool.QueryRow(ctx, query, id)
	participant := &Participant{}

	err := row.Scan(
		&participant.ID, &participant.SlotID, &participant.UserID,
		&participant.Status, &participant.ReservedAt, &participant.PaidAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}

	return participant, nil
}

// GetParticipantBySlotAndUser получает участие пользователя в слоте
func (d *DB) GetParticipantBySlotAndUser(ctx context.Context, slotID uuid.UUID, userID uuid.UUID) (*Participant, error) {
	query := `
		SELECT
			id, slot_id, user_id,
			status, reserved_at, paid_at
		FROM participants
		WHERE slot_id = $1 AND user_id = $2
		LIMIT 1
	`

	row := d.pool.QueryRow(ctx, query, slotID, userID)
	participant := &Participant{}

	err := row.Scan(
		&participant.ID, &participant.SlotID, &participant.UserID,
		&participant.Status, &participant.ReservedAt, &participant.PaidAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}

	return participant, nil
}

// CreateParticipant создает новое участие (INSERT)
// метод для присоединения к слоту
func (d *DB) CreateParticipant(ctx context.Context, arg *CreateParticipantParams) (*Participant, error) {
	query := `
		INSERT INTO participants (slot_id, user_id, status)
		VALUES ($1, $2, COALESCE($3, 'RESERVED'))
		RETURNING
			id, slot_id, user_id,
			status, reserved_at, paid_at
	`

	row := d.pool.QueryRow(ctx, query, arg.SlotID, arg.UserID, arg.Status)
	participant := &Participant{}

	err := row.Scan(
		&participant.ID, &participant.SlotID, &participant.UserID,
		&participant.Status, &participant.ReservedAt, &participant.PaidAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create participant: %w", err)
	}

	return participant, nil
}

// UpdateParticipantStatus обновляет статус участия
func (d *DB) UpdateParticipantStatus(ctx context.Context, id uuid.UUID, status string) (*Participant, error) {
	query := `
		UPDATE participants
		SET
			status = $2,
			paid_at = CASE WHEN $2 = 'PAID' THEN NOW() ELSE paid_at END
		WHERE id = $1
		RETURNING
			id, slot_id, user_id,
			status, reserved_at, paid_at
	`

	row := d.pool.QueryRow(ctx, query, id, status)
	participant := &Participant{}

	err := row.Scan(
		&participant.ID, &participant.SlotID, &participant.UserID,
		&participant.Status, &participant.ReservedAt, &participant.PaidAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update participant status: %w", err)
	}

	return participant, nil
}

// UpdateParticipantStatusWithCheck обновляет статус с проверкой текущего статуса
// КРИТИЧНО для защиты от двойной оплаты
func (d *DB) UpdateParticipantStatusWithCheck(ctx context.Context, id uuid.UUID, newStatus string, expectedStatus string) (*Participant, error) {
	query := `
		UPDATE participants
		SET
			status = $2,
			paid_at = CASE WHEN $2 = 'PAID' THEN NOW() ELSE paid_at END
		WHERE id = $1 AND status = $3
		RETURNING
			id, slot_id, user_id,
			status, reserved_at, paid_at
	`

	row := d.pool.QueryRow(ctx, query, id, newStatus, expectedStatus)
	participant := &Participant{}

	err := row.Scan(
		&participant.ID, &participant.SlotID, &participant.UserID,
		&participant.Status, &participant.ReservedAt, &participant.PaidAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Статус не совпадал или участие не найдено
		}
		return nil, fmt.Errorf("failed to update participant status: %w", err)
	}

	return participant, nil
}

// ListParticipantsByUser получает все участия пользователя
func (d *DB) ListParticipantsByUser(ctx context.Context, userID uuid.UUID) ([]Participant, error) {
	query := `
		SELECT
			id, slot_id, user_id,
			status, reserved_at, paid_at
		FROM participants
		WHERE user_id = $1
		ORDER BY reserved_at DESC
	`

	rows, err := d.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list participants: %w", err)
	}
	defer rows.Close()

	participants := []Participant{}
	for rows.Next() {
		participant := Participant{}
		err := rows.Scan(
			&participant.ID, &participant.SlotID, &participant.UserID,
			&participant.Status, &participant.ReservedAt, &participant.PaidAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}
		participants = append(participants, participant)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return participants, nil
}

// GetUser получает пользователя по ID
func (d *DB) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	row := d.pool.QueryRow(ctx, query, id)
	user := &User{}

	err := row.Scan(
		&user.ID, &user.Name, &user.Email,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// ListUsers получает всех пользователей
func (d *DB) ListUsers(ctx context.Context) ([]User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := d.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		user := User{}
		err := rows.Scan(
			&user.ID, &user.Name, &user.Email,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}

// GetPayment получает платеж по ID
func (d *DB) GetPayment(ctx context.Context, id uuid.UUID) (*Payment, error) {
	query := `
		SELECT id, participant_id, amount, status, idempotency_key, created_at, updated_at
		FROM payments
		WHERE id = $1
		LIMIT 1
	`

	row := d.pool.QueryRow(ctx, query, id)
	payment := &Payment{}

	err := row.Scan(
		&payment.ID, &payment.ParticipantID, &payment.Amount, &payment.Status, &payment.IdempotencyKey,
		&payment.CreatedAt, &payment.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return payment, nil
}

// CreatePayment создает новый платеж
func (d *DB) CreatePayment(ctx context.Context, arg *CreatePaymentParams) (*Payment, error) {
	query := `
		INSERT INTO payments (participant_id, amount, status, idempotency_key)
		VALUES ($1, $2, $3, $4)
		RETURNING id, participant_id, amount, status, idempotency_key, created_at, updated_at
	`

	row := d.pool.QueryRow(ctx, query, arg.ParticipantID, arg.Amount, arg.Status, arg.IdempotencyKey)
	payment := &Payment{}

	err := row.Scan(
		&payment.ID, &payment.ParticipantID, &payment.Amount, &payment.Status, &payment.IdempotencyKey,
		&payment.CreatedAt, &payment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}

// UpdatePaymentStatus обновляет статус платежа
func (d *DB) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status string) (*Payment, error) {
	query := `
		UPDATE payments
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, participant_id, amount, status, idempotency_key, created_at, updated_at
	`

	row := d.pool.QueryRow(ctx, query, id, status)
	payment := &Payment{}

	err := row.Scan(
		&payment.ID, &payment.ParticipantID, &payment.Amount, &payment.Status, &payment.IdempotencyKey,
		&payment.CreatedAt, &payment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	return payment, nil
}

// GetPaymentByIdempotencyKey получает платеж по ключу идемпотентности
func (d *DB) GetPaymentByIdempotencyKey(ctx context.Context, key string) (*Payment, error) {
	query := `
		SELECT id, participant_id, amount, status, idempotency_key, created_at, updated_at
		FROM payments
		WHERE idempotency_key = $1
		LIMIT 1
	`

	row := d.pool.QueryRow(ctx, query, key)
	payment := &Payment{}

	err := row.Scan(
		&payment.ID, &payment.ParticipantID, &payment.Amount, &payment.Status, &payment.IdempotencyKey,
		&payment.CreatedAt, &payment.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment by idempotency key: %w", err)
	}

	return payment, nil
}

package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ChrolloLucii/SNKI/apps/api/internal/locking"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)


type JoinRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

// Join handles POST /slots/{id}/join
func Join(pool *pgxpool.Pool, locker locking.Locker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		slotIDStr := chi.URLParam(r, "slotId")
		slotID, _ := uuid.Parse(slotIDStr)

		userIDStr := GetUserID(ctx)
		req := JoinRequest{UserID: uuid.MustParse(userIDStr)}
		if req.UserID == uuid.Nil {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusBadRequest, "missing_user_id", "INVALID_INPUT", "")
			return
		}

		lockKey := fmt.Sprintf("slot:%s:join", slotID)
		lock, err := locker.TryAcquire(ctx, lockKey, 5*time.Second)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusConflict, "slot_busy", "SLOT_BUSY", "")
			return
		}
		defer lock.Release(ctx)

		// Check slot exists
		var status string
		var capacity int32
		var deadline time.Time
		err = pool.QueryRow(ctx, `
			SELECT status, capacity, deadline_at FROM slots WHERE id = $1
		`, slotID).Scan(&status, &capacity, &deadline)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusNotFound, "slot_not_found", "NOT_FOUND", "")
			return
		}

		if status != "OPEN" {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusBadRequest, "slot_not_open", "INVALID_SLOT_STATUS", "")
			return
		}

		if time.Now().After(deadline) {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusBadRequest, "deadline_passed", "REGISTRATION_CLOSED", "")
			return
		}

		// Check not already joined
		var count int
		pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM participants WHERE slot_id = $1 AND user_id = $2
		`, slotID, req.UserID).Scan(&count)

		if count > 0 {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusConflict, "already_joined", "DUPLICATE_PARTICIPATION", "")
			return
		}

		// Count current
		count = 0
		pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM participants
			WHERE slot_id = $1 AND status IN ('RESERVED', 'PAID')
		`, slotID).Scan(&count)

		if int32(count) >= capacity {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusConflict, "slot_full", "OVERBOOKING", fmt.Sprintf("Slot is full (%d/%d)", count, capacity))
			return
		}

		// INSERT in transaction (create user if needed, then add participant)
		tx, _ := pool.Begin(ctx)
		defer tx.Rollback(ctx)

		// Create user if not exists (demo users don't have tokens)
		tx.Exec(ctx, `
			INSERT INTO users (id, token) 
			VALUES ($1, $2) 
			ON CONFLICT (id) DO NOTHING
		`, req.UserID, fmt.Sprintf("demo_%s", req.UserID))

		// Add participant
		tx.Exec(ctx, `
			INSERT INTO participants (slot_id, user_id, status)
			VALUES ($1, $2, 'RESERVED')
		`, slotID, req.UserID)

		tx.Commit(ctx)

		log.Printf("✓ User %s joined slot (now %d/%d)", req.UserID, count+1, capacity)

		w.Header().Set("Content-Type", "application/json")
		WriteJSON(w, http.StatusCreated, SuccessResp{Success: true, Message: fmt.Sprintf("Joined slot (%d/%d participants)", count+1, capacity)})
	}
}

// GetSlotInf handles GET /slots/{id}
func GetSlotInf(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slotIDStr := chi.URLParam(r, "slotId")
		slotID, _ := uuid.Parse(slotIDStr)

		var id, sport string
		var capacity int32
		var participants int
		
		err := pool.QueryRow(r.Context(), `
			SELECT 
				s.id, 
				s.sport, 
				s.capacity,
				COALESCE(COUNT(p.id), 0) as participants
			FROM slots s
			LEFT JOIN participants p ON s.id = p.slot_id AND p.status IN ('RESERVED', 'PAID')
			WHERE s.id = $1
			GROUP BY s.id, s.sport, s.capacity
		`, slotID).Scan(&id, &sport, &capacity, &participants)

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusNotFound, "not_found", "NOT_FOUND", "")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		WriteJSON(w, http.StatusOK, SuccessResp{
			Success: true, 
			Data: map[string]interface{}{
				"id": id, 
				"sport": sport,
				"capacity": capacity,
				"participants": participants,
			},
		})
	}
}

// ListAll handles GET /slots
func ListAll(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := pool.Query(r.Context(), `
			SELECT id, sport FROM slots WHERE status = 'OPEN' LIMIT 100
		`)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusInternalServerError, "error", "INTERNAL_ERROR", "")
			return
		}
		defer rows.Close()

		var slots []interface{}
		for rows.Next() {
			var id, sport string
			if err := rows.Scan(&id, &sport); err == nil {
				slots = append(slots, map[string]string{"id": id, "sport": sport})
			}
		}

		w.Header().Set("Content-Type", "application/json")
		WriteJSON(w, http.StatusOK, SuccessResp{Success: true, Data: slots})
	}
}

type PayRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Amount int       `json:"amount"`
}

// Pay handles POST /slots/{id}/pay
func Pay(pool *pgxpool.Pool, locker locking.Locker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		slotIDStr := chi.URLParam(r, "slotId")
		slotID, _ := uuid.Parse(slotIDStr)

		idempotencyKey := r.Header.Get("X-Idempotency-Key")
		if idempotencyKey == "" {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusBadRequest, "missing_idempotency_key", "INVALID_INPUT", "")
			return
		}

		userIDStr := GetUserID(ctx)
		req := PayRequest{}
		json.NewDecoder(r.Body).Decode(&req)
		req.UserID = uuid.MustParse(userIDStr)

		if req.UserID == uuid.Nil || req.Amount <= 0 {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusBadRequest, "invalid_input", "INVALID_INPUT", "")
			return
		}

		// существует ли уже ключ идемпотентности
		var prevStatus string
		err := pool.QueryRow(ctx, `
			SELECT status FROM payments WHERE idempotency_key = $1
		`, idempotencyKey).Scan(&prevStatus)
		
		if err == nil {
			// Idempotency hit
			w.Header().Set("Content-Type", "application/json")
			WriteJSON(w, http.StatusOK, SuccessResp{
				Success: true, 
				Message: fmt.Sprintf("Payment already processed (idempotent req)"),
			})
			return
		}

		// Получиние распределенной блокировки для оплаты
		lockKey := fmt.Sprintf("payment:participant:%s:%s", req.UserID, slotID)
		lock, err := locker.TryAcquire(ctx, lockKey, 10*time.Second)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusConflict, "payment_in_progress", "PAYMENT_LOCKED", "")
			return
		}
		defer lock.Release(ctx)

		//  Проверка статуса участника и защита от двойной оплаты в рамках транзакции
		tx, err := pool.Begin(ctx)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "internal_error", "INTERNAL_ERROR", "internal server error")
			return
		}
		defer tx.Rollback(ctx)

		// Получаем participant_id и текущий статус участника для данного слота и пользователя
		var participantID string
		var status string
		
		err = tx.QueryRow(ctx, `
			SELECT id, status FROM participants 
			WHERE slot_id = $1 AND user_id = $2
		`, slotID, req.UserID).Scan(&participantID, &status)

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusNotFound, "participant_not_found", "NOT_FOUND", "")
			return
		}

		// Если статус уже PAID, то возвращаем 409 Conflict (двойная оплата)
		if status == "PAID" {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusConflict, "already_paid", "ALREADY_PAID", "Participant has already paid for this slot")
			return
		}

		if status != "RESERVED" {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusBadRequest, "invalid_status", "INVALID_STATUS", "")
			return
		}

		// INSERT платежа (или вернуть ошибку, если idempotency_key уже существует)
		_, err = tx.Exec(ctx, `
			INSERT INTO payments (participant_id, idempotency_key, amount, status)
			VALUES ($1, $2, $3, 'PAID')
		`, participantID, idempotencyKey, req.Amount)

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict) // 409 Conflict для двойной оплаты
			WriteJSON(w, http.StatusOK, ErrorResp{Error: "payment_failed", Code: "DOUBLE_PAYMENT_DB", Message: err.Error()})
			return
		}

		// Обновляем статус участника на PAID
		_, err = tx.Exec(ctx, `
			UPDATE participants SET status = 'PAID' WHERE id = $1
		`, participantID)

		if err != nil {
			WriteError(w, http.StatusInternalServerError, "internal_error", "INTERNAL_ERROR", "failed to update participant")
			return
		}

		tx.Commit(ctx)

		w.Header().Set("Content-Type", "application/json")
		WriteJSON(w, http.StatusOK, SuccessResp{Success: true, Message: "Payment successful"})
	}
}



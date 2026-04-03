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
	UserID string `json:"user_id"`
}

// Join handles POST /slots/{id}/join
func Join(pool *pgxpool.Pool, locker locking.Locker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		slotIDStr := chi.URLParam(r, "slotId")
		slotID, _ := uuid.Parse(slotIDStr)

		userIDStr := GetUserID(ctx)
		req := JoinRequest{UserID: userIDStr}
		if req.UserID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResp{Error: "missing_user_id", Code: "INVALID_INPUT"})
			return
		}

		lockKey := fmt.Sprintf("slot:%s:join", slotID)
		lock, err := locker.TryAcquire(ctx, lockKey, 5*time.Second)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(ErrorResp{Error: "slot_busy", Code: "SLOT_BUSY"})
			return
		}
		defer lock.Release(ctx)

		// In our DB `users` table, ID is a UUID. If they provide demo_user, we need to hash it 
		// into a valid UUID or generate a deterministic one. Let's use name-based uuid.
		userUUID := uuid.NewMD5(uuid.NameSpaceDNS, []byte(req.UserID))

		//  проверяем, что слот существует, открыт для регистрации и не заполнен, чтобы избежать гонок и обеспечить целостность данных (чтобы всегда был актуальный статус слота при попытке записаться)
		var status string
		var capacity int32
		var deadline time.Time
		err = pool.QueryRow(ctx, `
			SELECT status, capacity, deadline_at FROM slots WHERE id = $1
		`, slotID).Scan(&status, &capacity, &deadline)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResp{Error: "slot_not_found", Code: "NOT_FOUND"})
			return
		}

		if status != "OPEN" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResp{Error: "slot_not_open", Code: "INVALID_SLOT_STATUS"})
			return
		}

		if time.Now().After(deadline) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResp{Error: "deadline_passed", Code: "REGISTRATION_CLOSED"})
			return
		}

		// проверяем, что пользователь еще не записан в этот слот, чтобы избежать дублей и обеспечить idempotent join (чтобы при повторных запросах с тем же userID не создавались новые участники, а просто возвращалась ошибка о том, что пользователь уже записан)
		var count int
		pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM participants WHERE slot_id = $1 AND user_id = $2
		`, slotID, userUUID).Scan(&count)

		if count > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(ErrorResp{Error: "already_joined", Code: "DUPLICATE_PARTICIPATION"})
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
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(ErrorResp{Error: "slot_full", Code: "OVERBOOKING", Message: fmt.Sprintf("Slot is full (%d/%d)", count, capacity)})
			return
		}

		// INSERT в participants и создание пользователя, если его нет, нужно делать в рамках одной транзакции, чтобы избежать гонок и обеспечить целостность данных (чтобы всегда был пользователь для участника)
		tx, _ := pool.Begin(ctx)
		defer tx.Rollback(ctx)

		// создаем пользователя, если его нет, чтобы обеспечить целостность данных (чтобы всегда был пользователь для участника)
		tx.Exec(ctx, `
			INSERT INTO users (id, token) 
			VALUES ($1, $2) 
			ON CONFLICT (id) DO NOTHING
		`, userUUID, req.UserID)

		// Add participant
		tx.Exec(ctx, `
			INSERT INTO participants (slot_id, user_id, status)
			VALUES ($1, $2, 'RESERVED')
		`, slotID, userUUID)

		tx.Commit(ctx)

		log.Printf("✓ User %s joined slot (now %d/%d)", userUUID, count+1, capacity)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(SuccessResp{Success: true, Message: fmt.Sprintf("Joined slot (%d/%d participants)", count+1, capacity)})
	}
}

//  GET /slots/{id} - получить базовую информацию о слоте (спорт, район, название площадки, количество участников и т.д.)
func GetSlotInf(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slotIDStr := chi.URLParam(r, "slotId")
		slotID, _ := uuid.Parse(slotIDStr)

		var id, sport, district, venueName, address, status string
		var capacity, durationMinutes, expectedPrice, maxPrice int
		var participants int
		var startsAt, deadlineAt time.Time
		var rulesText *string
		
		err := pool.QueryRow(r.Context(), `
			SELECT 
				s.id, s.sport, s.district, s.venue_name, s.address,
				s.starts_at, s.deadline_at, s.duration_minutes,
				s.capacity, s.expected_price, s.max_price, s.rules_text, s.status,
				COALESCE(COUNT(p.id), 0) as participants
			FROM slots s
			LEFT JOIN participants p ON s.id = p.slot_id AND p.status IN ('RESERVED', 'PAID')
			WHERE s.id = $1
			GROUP BY s.id
		`, slotID).Scan(&id, &sport, &district, &venueName, &address,
			&startsAt, &deadlineAt, &durationMinutes,
			&capacity, &expectedPrice, &maxPrice, &rulesText, &status, &participants)

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResp{Error: "not_found", Code: "NOT_FOUND"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SuccessResp{
			Success: true, 
			Data: map[string]interface{}{
				"id": id, 
				"sport": sport,
				"district": district,
				"venue_name": venueName,
				"address": address,
				"starts_at": startsAt.Format(time.RFC3339),
				"deadline_at": deadlineAt.Format(time.RFC3339),
				"duration_minutes": durationMinutes,
				"capacity": capacity,
				"expected_price": expectedPrice,
				"max_price": maxPrice,
				"rules_text": rulesText,
				"status": status,
				"participants": participants,
				"free_spots": capacity - participants,
			},
		})
	}
}

// ListAll handles GET /slots
func ListAll(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sportFilter := r.URL.Query().Get("sport")
		districtFilter := r.URL.Query().Get("district")

		query := `
			SELECT 
				s.id, s.sport, s.district, s.venue_name, s.address,
				s.starts_at, s.deadline_at, s.duration_minutes,
				s.capacity, s.expected_price, s.max_price, s.rules_text, s.status,
				COALESCE(COUNT(p.id), 0) as participants
			FROM slots s
			LEFT JOIN participants p ON s.id = p.slot_id AND p.status IN ('RESERVED', 'PAID')
			WHERE s.status = 'OPEN' 
		`
		
		args := []interface{}{}
		argCount := 1
		
		if sportFilter != "" {
			query += fmt.Sprintf(" AND s.sport = $%d", argCount)
			args = append(args, sportFilter)
			argCount++
		}
		
		if districtFilter != "" {
			query += fmt.Sprintf(" AND s.district ILIKE $%d", argCount)
			args = append(args, "%"+districtFilter+"%")
			argCount++
		}

		query += ` GROUP BY s.id ORDER BY s.starts_at ASC LIMIT 100`

		rows, err := pool.Query(r.Context(), query, args...)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResp{Error: "error", Code: "INTERNAL_ERROR"})
			return
		}
		defer rows.Close()

		var slots []map[string]interface{}
		for rows.Next() {
			var id, sport, district, venueName, address, status string
			var capacity, durationMinutes, expectedPrice, maxPrice int
			var participants int
			var startsAt, deadlineAt time.Time
			var rulesText *string
			
			if err := rows.Scan(&id, &sport, &district, &venueName, &address,
				&startsAt, &deadlineAt, &durationMinutes,
				&capacity, &expectedPrice, &maxPrice, &rulesText, &status, &participants); err == nil {
				
				slots = append(slots, map[string]interface{}{
					"id": id, 
					"sport": sport,
					"district": district,
					"venue_name": venueName,
					"address": address,
					"starts_at": startsAt.Format(time.RFC3339),
					"deadline_at": deadlineAt.Format(time.RFC3339),
					"duration_minutes": durationMinutes,
					"capacity": capacity,
					"expected_price": expectedPrice,
					"max_price": maxPrice,
					"rules_text": rulesText,
					"status": status,
					"participants": participants,
					"free_spots": capacity - participants,
				})
			}
		}
		
		if slots == nil {
			slots = []map[string]interface{}{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SuccessResp{Success: true, Data: slots})
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
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResp{Error: "missing_idempotency_key", Code: "INVALID_INPUT"})
			return
		}

		userIDStr := GetUserID(ctx)
		userUUID := uuid.NewMD5(uuid.NameSpaceDNS, []byte(userIDStr))
		
		req := PayRequest{}
		json.NewDecoder(r.Body).Decode(&req)
		// Map token string to UUID.
		req.UserID = userUUID
		
		if req.UserID == uuid.Nil || req.Amount <= 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResp{Error: "invalid_input", Code: "INVALID_INPUT"})
			return
		}

		// Проверяем, не был ли уже обработан запрос с таким idempotency key
		var prevStatus string
		err := pool.QueryRow(ctx, `
			SELECT status FROM payments WHERE idempotency_key = $1
		`, idempotencyKey).Scan(&prevStatus)
		
		if err == nil {
			// Если запись найдена, значит запрос уже обрабатывался
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(SuccessResp{
				Success: true, 
				Message: fmt.Sprintf("Payment already processed (idempotent req)"),
			})
			return
		}

		// здесь можно было бы еще дополнительно проверять, что ошибка именно "no rows", а не какая-то другая, но для простоты примера опустим это
		lockKey := fmt.Sprintf("payment:participant:%s:%s", req.UserID, slotID)
		lock, err := locker.TryAcquire(ctx, lockKey, 10*time.Second)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(ErrorResp{Error: "payment_in_progress", Code: "PAYMENT_LOCKED"})
			return
		}
		defer lock.Release(ctx)

		tx, err := pool.Begin(ctx)
		if err != nil {
			http.Error(w, "internal server error", 500)
			return
		}
		defer tx.Rollback(ctx)
		// Проверяем, что участник существует и в правильном статусе для оплаты
		var participantID string
		var status string
		
		err = tx.QueryRow(ctx, `
			SELECT id, status FROM participants 
			WHERE slot_id = $1 AND user_id = $2
		`, slotID, req.UserID).Scan(&participantID, &status)

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResp{Error: "participant_not_found", Code: "NOT_FOUND"})
			return
		}

		//  409 конфликт, если уже оплачено (можно было бы и 400, но так понятнее)
		if status == "PAID" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(ErrorResp{Error: "already_paid", Code: "ALREADY_PAID", Message: "Participant has already paid for this slot"})
			return
		}

		if status != "RESERVED" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResp{Error: "invalid_status", Code: "INVALID_STATUS"})
			return
		}

		// INSERT 
		_, err = tx.Exec(ctx, `
			INSERT INTO payments (participant_id, idempotency_key, amount, status)
			VALUES ($1, $2, $3, 'PAID')
		`, participantID, idempotencyKey, req.Amount)

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict) 
			json.NewEncoder(w).Encode(ErrorResp{Error: "payment_failed", Code: "DOUBLE_PAYMENT_DB", Message: err.Error()})
			return
		}

		_, err = tx.Exec(ctx, `
			UPDATE participants SET status = 'PAID' WHERE id = $1
		`, participantID)

		if err != nil {
			http.Error(w, "failed to update participant", 500)
			return
		}

		tx.Commit(ctx)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SuccessResp{Success: true, Message: "Payment successful"})
	}
}



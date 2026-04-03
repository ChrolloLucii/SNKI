package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ParticipationSlotItem struct {
	// Participation Data
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Status     string    `json:"status"`
	ReservedAt time.Time `json:"reserved_at"`
	PaidAt     *time.Time `json:"paid_at,omitempty"`

	// Slot Data
	SlotID          string    `json:"slot_id"`
	Sport           string    `json:"sport"`
	District        string    `json:"district"`
	VenueName       string    `json:"venue_name"`
	Address         string    `json:"address"`
	StartsAt        time.Time `json:"starts_at"`
	DeadlineAt      time.Time `json:"deadline_at"`
	DurationMinutes int       `json:"duration_minutes"`
	ExpectedPrice   int       `json:"expected_price"`
	MaxPrice        int       `json:"max_price"`
	SlotStatus      string    `json:"slot_status"`
}

func GetMyParticipations(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Temporary authentication via headers directly, before middleware is used everywhere
		token := r.Header.Get("X-Demo-Token")
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResp{Error: "unauthorized", Code: "UNAUTHORIZED", Message: "Missing X-Demo-Token header"})
			return
		}

		ctx := r.Context()
		
		// For demo MVP, assume the token might literally be the user UUID.
		// If not a valid UUID format, fail.
		userID := token

		query := `
			SELECT 
				p.id, p.user_id, p.status, p.reserved_at, p.paid_at,
				s.id as slot_id, s.sport, s.district, s.venue_name, s.address,
				s.starts_at, s.deadline_at, s.duration_minutes, s.expected_price, s.max_price, s.status as slot_status
			FROM participants p
			JOIN slots s ON p.slot_id = s.id
			WHERE p.user_id = $1
			ORDER BY s.starts_at DESC
		`

		rows, err := pool.Query(ctx, query, userID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResp{Error: "db_error", Code: "INTERNAL_ERROR", Message: err.Error()})
			return
		}
		defer rows.Close()

		var results []ParticipationSlotItem
		for rows.Next() {
			var i ParticipationSlotItem
			if err := rows.Scan(
				&i.ID, &i.UserID, &i.Status, &i.ReservedAt, &i.PaidAt,
				&i.SlotID, &i.Sport, &i.District, &i.VenueName, &i.Address,
				&i.StartsAt, &i.DeadlineAt, &i.DurationMinutes, &i.ExpectedPrice, &i.MaxPrice, &i.SlotStatus,
			); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResp{Error: "db_error", Code: "INTERNAL_ERROR", Message: err.Error()})
				return
			}
			results = append(results, i)
		}

		if results == nil {
			results = []ParticipationSlotItem{} // Return empty array instead of null
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(results)
	}
}

package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminCreateSlotRequest struct {
	Sport           string    `json:"sport"`
	District        string    `json:"district"`
	VenueName       string    `json:"venue_name"`
	Address         string    `json:"address"`
	StartsAt        time.Time `json:"starts_at"`
	DeadlineAt      time.Time `json:"deadline_at"`
	DurationMinutes int       `json:"duration_minutes"`
	Capacity        int       `json:"capacity"`
	MinPlayers      int       `json:"min_players"`
	ExpectedPrice   int       `json:"expected_price"`
	MaxPrice        int       `json:"max_price"`
	RulesText       string    `json:"rules_text,omitempty"`
	Status          string    `json:"status,omitempty"`
}

type SlotResp struct {
	ID              string    `json:"id"`
	Sport           string    `json:"sport"`
	District        string    `json:"district"`
	VenueName       string    `json:"venue_name"`
	Address         string    `json:"address"`
	StartsAt        time.Time `json:"starts_at"`
	DeadlineAt      time.Time `json:"deadline_at"`
	DurationMinutes int       `json:"duration_minutes"`
	Capacity        int       `json:"capacity"`
	MinPlayers      int       `json:"min_players"`
	ExpectedPrice   int       `json:"expected_price"`
	MaxPrice        int       `json:"max_price"`
	RulesText       *string   `json:"rules_text,omitempty"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func CreateSlot(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Temporary authentication
		token := r.Header.Get("X-Admin-Token")
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED", "Missing X-Admin-Token header")
			return
		}

		var req AdminCreateSlotRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusUnprocessableEntity, "invalid_input", "VALIDATION_ERROR", "Invalid JSON format")
			return
		}

		if req.Sport == "" || req.District == "" || req.VenueName == "" || req.Address == "" {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusUnprocessableEntity, "missing_fields", "VALIDATION_ERROR", "Missing required text fields")
			return
		}

		if req.DurationMinutes <= 0 || req.Capacity <= 0 || req.MinPlayers <= 0 {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusUnprocessableEntity, "invalid_numbers", "VALIDATION_ERROR", "Numbers must be positive")
			return
		}

		status := req.Status
		if status == "" {
			status = "OPEN"
		}

		query := `
			INSERT INTO slots (
				sport, district, venue_name, address, starts_at, deadline_at, duration_minutes,
				capacity, min_players, expected_price, max_price, rules_text, status
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
			) RETURNING id, sport, district, venue_name, address, starts_at, deadline_at, duration_minutes, capacity, min_players, expected_price, max_price, rules_text, status, created_at, updated_at
		`

		var s SlotResp
		err := pool.QueryRow(r.Context(), query,
			req.Sport, req.District, req.VenueName, req.Address, req.StartsAt, req.DeadlineAt,
			req.DurationMinutes, req.Capacity, req.MinPlayers, req.ExpectedPrice, req.MaxPrice, req.RulesText, status,
		).Scan(
			&s.ID, &s.Sport, &s.District, &s.VenueName, &s.Address, &s.StartsAt, &s.DeadlineAt,
			&s.DurationMinutes, &s.Capacity, &s.MinPlayers, &s.ExpectedPrice, &s.MaxPrice, &s.RulesText, &s.Status, &s.CreatedAt, &s.UpdatedAt,
		)

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			WriteError(w, http.StatusInternalServerError, "db_error", "INTERNAL_ERROR", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		WriteJSON(w, http.StatusOK, s)
	}
}

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
)

type contextKey string

const userIDKey contextKey = "userId"

// AuthUserMiddleware verifies the X-Demo-Token header and puts the userID in context.
func AuthUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Demo-Token")
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResp{Error: "unauthorized", Code: "UNAUTHORIZED", Message: "Missing X-Demo-Token header"})
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthAdminMiddleware verifies the X-Admin-Token header for admin actions.
func AuthAdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Admin-Token")
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResp{Error: "unauthorized", Code: "UNAUTHORIZED", Message: "Missing X-Admin-Token header"})
			return
		}

		// Simple demo token check
		// Could add actual token validation logic here if required
		next.ServeHTTP(w, r)
	})
}

// GetUserID extracts the user ID from the request context.
func GetUserID(ctx context.Context) string {
	val, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return ""
	}
	return val
}

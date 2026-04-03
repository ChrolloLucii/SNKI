package handlers

import (
	"context"
	"encoding/json"
	"net/http"
)

type contextKey string

const userIDKey contextKey = "userId"

// AuthUserMiddleware - это промежуточное ПО, которое извлекает userID из заголовков запроса и сохраняет его в контексте для дальнейшего использования в обработчиках. В реальной системе здесь будет полноценная аутентификация, например, через JWT или сессию.
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
// AuthAdminMiddleware проверяет наличие X-Admin-Token в заголовках и возвращает 401, если его нет. В реальной реализации здесь могла бы быть более сложная логика проверки токена.
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

// GetUserID возвращает userID из контекста, установленного в AuthUserMiddleware. Если userID нет, возвращает пустую строку.
func GetUserID(ctx context.Context) string {
	val, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return ""
	}
	return val
}

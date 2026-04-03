package handlers

import (
	"encoding/json"
	"net/http"
)

type ErrorResp struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

type SuccessResp struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// WriteJSON sends a JSON response with the specified status code
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// WriteError formats and sends an error response using WriteJSON
func WriteError(w http.ResponseWriter, status int, errorKey string, code string, message string) {
	WriteJSON(w, status, ErrorResp{
		Error:   errorKey,
		Code:    code,
		Message: message,
	})
}

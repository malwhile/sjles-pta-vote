package common

import (
	"encoding/json"
	"net/http"
)

const (
	DATE_FORMAT = "2006-01-02 15:04:05"
	SUCCESS = "success"
)

// StandardResponse is the standard API response format for all endpoints
type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SendSuccess sends a successful response with data
func SendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(StandardResponse{
		Success: true,
		Data:    data,
	})
}

// SendError sends an error response with status code
func SendError(w http.ResponseWriter, errStr string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(StandardResponse{
		Success: false,
		Error:   errStr,
	})
}

// SendErrorWithCode sends an error response with detailed error information
func SendErrorWithCode(w http.ResponseWriter, message string, code string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
		"code":    code,
	})
}

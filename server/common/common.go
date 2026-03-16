package common

import (
	"encoding/json"
	"net/http"
)

const (
	DATE_FORMAT = "2006-01-02 15:04:05"
	SUCCESS = "success"
)

func SendError(w http.ResponseWriter, errStr string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": errStr})
}

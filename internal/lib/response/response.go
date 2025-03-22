package response

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
}

func Error(w http.ResponseWriter, statusCode int, message string) {
	type errResponse struct {
		Error string `json:"error"`
	}
	JSON(w, statusCode, errResponse{Error: message})
}

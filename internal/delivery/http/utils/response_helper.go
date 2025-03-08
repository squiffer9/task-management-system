package utils

import (
	"encoding/json"
	"net/http"
)

// ResponseWrapper standardizes API responses
type ResponseWrapper struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo provides detailed error information
type ErrorInfo struct {
	Code    int    `json:"code" example:"404"`
	Message string `json:"message" example:"Resource not found"`
}

// RespondWithError sends an error response in a standardized format
func RespondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := ResponseWrapper{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// RespondWithJSON sends a success response in a standardized format
func RespondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := ResponseWrapper{
		Success: true,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

// RespondWithJSONDirect sends the raw data as JSON response (without the wrapper)
// Useful for backward compatibility with existing clients
func RespondWithJSONDirect(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

package server

import (
	"encoding/json"
	"net/http"
)

// Error response structure
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details any    `json:"details,omitempty"`
}

// JSON writes a JSON response
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// Error writes an error response
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, ErrorResponse{Error: msg})
}

// ErrorWithCode writes an error response with code
func ErrorWithCode(w http.ResponseWriter, status int, code, msg string) {
	JSON(w, status, ErrorResponse{Error: msg, Code: code})
}

// Created writes a 201 response with Location header
func Created(w http.ResponseWriter, location string, data any) {
	if location != "" {
		w.Header().Set("Location", location)
	}
	JSON(w, http.StatusCreated, data)
}

// NoContent writes a 204 response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Decode decodes JSON request body
func Decode(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

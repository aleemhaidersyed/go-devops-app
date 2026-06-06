package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse defines the structure of our health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Service   string `json:"service"`
}

// HealthHandler responds to GET /health
// It tells the caller the service is alive and running
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Build the response object
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service:   "go-devops-app",
	}

	// Set Content-Type header to tell the client we're sending JSON
	w.Header().Set("Content-Type", "application/json")

	// Write HTTP 200 OK status code
	w.WriteHeader(http.StatusOK)

	// Encode the response struct as JSON and write it to the response body
	json.NewEncoder(w).Encode(response)
}

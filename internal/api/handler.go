package api

import (
	"encoding/json"
	"net/http"
)

type DownloadRequest struct {
	URL string `json:"url"`
}

type DownloadResponse struct {
	Message string `json:"message"`
	URL    string `json:"url"`
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DownloadRequest

	// Decode the JSON request body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation for the URL
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Response 
	resp := DownloadResponse{
		Message: "Download started",
		URL:    req.URL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
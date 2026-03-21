package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"video-downloader/internal/downloader"
)

type DownloadRequest struct {
	URL string `json:"url"`
}

type DownloadResponse struct {
	Message string `json:"message"`
	File    string `json:"file"`
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DownloadRequest

	// Decode the JSON request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation for the URL
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Call downloader
	filePath, err := downloader.DownloadVideo(req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer os.Remove(filePath)

	file, err := os.Open(filePath)
    if err != nil {
        http.Error(w, "failed to open file", http.StatusInternalServerError)
        return
	}
	
	defer file.Close()

	stat, err := file.Stat()
    if err != nil {
        http.Error(w, "failed to stat file", http.StatusInternalServerError)
        return
    }

	w.Header().Set("Content-Disposition", `attachment; filename="video.mp4"`)
    w.Header().Set("Content-Type", "video/mp4")
    w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	if _, err = io.Copy(w, file); err != nil {
        fmt.Println("streaming error:", err)
	}
}
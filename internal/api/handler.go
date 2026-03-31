package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"video-downloader/internal/downloader"
)

func DownloadHandler(sem chan struct{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Step 1: Only allow POST requests
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Step 2: Decode JSON body
		var req struct {
			URL 	string `json:"url"`
			Format 	string `json:"format"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Step 3: Validate URL
		if req.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		sem <- struct{}{} // Acquire a slot — blocks here if 3 downloads are already running

		defer func() { <-sem}() // Release the slot when this function exits — no matter what happens
		
		// Step 4: Download the video
		filePath, err := downloader.DownloadVideo(req.URL, req.Format)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(filePath)

		// Step 5: Open the downloaded file
		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "Failed to open file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Step 6: Get file info for Content-Length header
		fileInfo, err := file.Stat()
		if err != nil {
			http.Error(w, "Failed to read file info", http.StatusInternalServerError)
			return
		}

		// Step 7: Set response headers
		fileName := filepath.Base(filePath)
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
		contentType := "video/mp4"
		if req.Format == "mp3" {
			contentType = "audio/mpeg"
		} else if req.Format == "webm" {
			contentType = "video/webm"
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

		// Step 8: Stream file to client
		_, err = io.Copy(w, file)
		if err != nil {
			fmt.Println("Error streaming file:", err)
		}

		// Step 9: Delete file from disk after sending
		// err = os.Remove(filePath)
		// if err != nil {
		// 	fmt.Println("Cleanup error:", err)
		// } else {
		// 	fmt.Println("Cleaned up:", filePath)
		// }
	}
}
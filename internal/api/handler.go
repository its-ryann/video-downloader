package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"video-downloader/internal/downloader"
)

// Job tracks the state of a single download
type Job struct {
	Status   string // "processing", "done", "error"
	Progress int    // 0-100
	FilePath string
	Error    string
}

// jobs stores all active downloads in memory
// sync.RWMutex protects it from concurrent access
var (
	jobs   = make(map[string]*Job)
	jobsMu sync.RWMutex
)

func setJob(id string, job *Job) {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	jobs[id] = job
}

func getJob(id string) (*Job, bool) {
	jobsMu.RLock()
	defer jobsMu.RUnlock()
	job, ok := jobs[id]
	return job, ok
}

func updateJob(id string, fn func(*Job)) {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	if job, ok := jobs[id]; ok {
		fn(job)
	}
}

// POST /download — starts a background job, returns job ID immediately
func StartDownload(sem chan struct{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			URL string `json:"url"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if req.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		sem <- struct{}{} // Acquire a slot — blocks here if 3 downloads are already running

		defer func() { <-sem}() // Release the slot when this function exits — no matter what happens
		
		// Step 4: Download the video
		filePath, err := downloader.DownloadVideo(req.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(filePath)

	file, err := os.Open(job.FilePath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Failed to read file info", http.StatusInternalServerError)
		return
	}

		// Step 7: Set response headers
		fileName := filepath.Base(filePath)
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	if _, err = io.Copy(w, file); err != nil {
		fmt.Println("Error streaming file:", err)
	}
}
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
			URL    string `json:"url"`
			Format string `json:"format"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if req.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}
		if req.Format == "" {
			req.Format = "mp4"
		}

		// Generate a unique job ID
		id := fmt.Sprintf("%d", time.Now().UnixNano())

		// Register the job
		setJob(id, &Job{Status: "processing", Progress: 0})

		// Run the download in the background
		go func() {
			sem <- struct{}{}
			defer func() { <-sem }()

			filePath, err := downloader.DownloadVideo(req.URL, req.Format, func(pct int) {
				updateJob(id, func(j *Job) {
					j.Progress = pct
				})
			})

			if err != nil {
				updateJob(id, func(j *Job) {
					j.Status = "error"
					j.Error = err.Error()
				})
				return
			}

			updateJob(id, func(j *Job) {
				j.Status = "done"
				j.Progress = 100
				j.FilePath = filePath
			})
		}()

		// Return job ID to client immediately
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"job_id": id})
	}
}

// GET /progress/{id} — returns current job status and progress
func GetProgress(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/progress/")
	job, ok := getJob(id)
	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":   job.Status,
		"progress": job.Progress,
		"error":    job.Error,
	})
}

// GET /file/{id} — serves the file once the job is done
func ServeFile(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/file/")
	job, ok := getJob(id)
	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	if job.Status != "done" {
		http.Error(w, "file not ready", http.StatusAccepted)
		return
	}

	defer os.Remove(job.FilePath)

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

	fileName := filepath.Base(job.FilePath)
	w.Header().Set("Content-Disposition", `attachment; filename="`+fileName+`"`)
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	if _, err = io.Copy(w, file); err != nil {
		fmt.Println("Error streaming file:", err)
	}
}
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"video-downloader/internal/downloader"
)

type Job struct {
	ID         string             `json:"id"`
	Status     string             `json:"status"` // "processing", "done", "error", "cancelled"
	Progress   int                `json:"progress"`
	FilePath   string             `json:"-"`
	Title      string             `json:"title"`
	Filename   string             `json:"filename"`
	Error      string             `json:"error"`
	CreatedAt  time.Time          `json:"created_at"`
	CancelFunc context.CancelFunc `json:"-"`
}

var (
	jobs   = make(map[string]*Job)
	jobsMu sync.RWMutex
	startTime = time.Now()
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

func deleteJob(id string) {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	if j, ok := jobs[id]; ok {
		if j.FilePath != "" {
			os.Remove(j.FilePath)
		}
		delete(jobs, id)
	}
}

// POST /info — extracts video metadata for frontend preview
func GetInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.URL) == "" {
		http.Error(w, "Valid URL is required", http.StatusBadRequest)
		return
	}

	info, err := downloader.GetVideoInfo(r.Context(), req.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// POST /download — starts a background download job
func StartDownload(sem chan struct{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			URL     string `json:"url"`
			Format  string `json:"format"`
			Quality string `json:"quality"`
			Title   string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.URL) == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}
		if req.Format == "" {
			req.Format = "mp4"
		}

		id := fmt.Sprintf("%d", time.Now().UnixNano())
		ctx, cancel := context.WithCancel(context.Background())

		job := &Job{
			ID:         id,
			Status:     "processing",
			Progress:   0,
			Title:      req.Title,
			CreatedAt:  time.Now(),
			CancelFunc: cancel,
		}
		setJob(id, job)

		go func() {
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				updateJob(id, func(j *Job) {
					j.Status = "cancelled"
					j.Error = "Job cancelled before starting"
				})
				return
			}

			filePath, ext, err := downloader.DownloadVideo(ctx, id, req.URL, req.Format, req.Quality, func(pct int) {
				updateJob(id, func(j *Job) {
					j.Progress = pct
				})
			})

			if err != nil {
				updateJob(id, func(j *Job) {
					if ctx.Err() == context.Canceled {
						j.Status = "cancelled"
						j.Error = "Download cancelled"
					} else {
						j.Status = "error"
						j.Error = err.Error()
					}
				})
				return
			}

			dispTitle := req.Title
			if dispTitle == "" {
				dispTitle = "download"
			}
			cleanName := downloader.SanitizeFilename(dispTitle) + ext

			updateJob(id, func(j *Job) {
				j.Status = "done"
				j.Progress = 100
				j.FilePath = filePath
				j.Filename = cleanName
			})
		}()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"job_id": id})
	}
}

// POST /cancel/{id} — cancels an active job
func CancelJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/cancel/")
	job, ok := getJob(id)
	if !ok {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	if job.CancelFunc != nil {
		job.CancelFunc()
	}

	updateJob(id, func(j *Job) {
		j.Status = "cancelled"
		j.Error = "Cancelled by user"
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cancelled", "job_id": id})
}

// GET /progress/{id} — returns job status
func GetProgress(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/progress/")
	job, ok := getJob(id)
	if !ok {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":   job.Status,
		"progress": job.Progress,
		"error":    job.Error,
		"filename": job.Filename,
	})
}

// GET /file/{id} — serves completed file with range support for video streaming
func ServeFile(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/file/")
	job, ok := getJob(id)
	if !ok {
		http.Error(w, "Job not found or expired", http.StatusNotFound)
		return
	}
	if job.Status != "done" {
		http.Error(w, "File not ready", http.StatusAccepted)
		return
	}

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

	ext := filepath.Ext(job.FilePath)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		switch ext {
		case ".mp3":
			contentType = "audio/mpeg"
		case ".mp4":
			contentType = "video/mp4"
		case ".webm":
			contentType = "video/webm"
		case ".mkv":
			contentType = "video/x-matroska"
		default:
			contentType = "application/octet-stream"
		}
	}

	downloadName := job.Filename
	if downloadName == "" {
		downloadName = filepath.Base(job.FilePath)
	}

	// Set disposition: attachment for explicit download, inline for video streaming
	if r.URL.Query().Get("dl") == "1" {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, downloadName))
	} else {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, downloadName))
	}
	w.Header().Set("Content-Type", contentType)

	// Use http.ServeContent to support HTTP Range requests for video seeking/streaming
	http.ServeContent(w, r, downloadName, fileInfo.ModTime(), file)
}

// GET /health — server health check
func GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"uptime": time.Since(startTime).String(),
	})
}

// StartJobCleaner — purges jobs and temp files older than maxAge
func StartJobCleaner(interval, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			now := time.Now()
			jobsMu.Lock()
			for id, j := range jobs {
				if now.Sub(j.CreatedAt) > maxAge {
					if j.CancelFunc != nil {
						j.CancelFunc()
					}
					if j.FilePath != "" {
						os.Remove(j.FilePath)
					}
					delete(jobs, id)
				}
			}
			jobsMu.Unlock()

			// Clean up orphaned vidsnap files in temp dir
			matches, _ := filepath.Glob(filepath.Join(os.TempDir(), "vidsnap_*"))
			for _, m := range matches {
				info, err := os.Stat(m)
				if err == nil && now.Sub(info.ModTime()) > maxAge {
					os.Remove(m)
				}
			}
		}
	}()
}
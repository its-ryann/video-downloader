package server

import (
	"fmt"
	"net/http"
	"os"
	"time"
	"video-downloader/internal/api"
)

func Start() {
	sem := make(chan struct{}, 3)

	// Start background job cleaner (checks every 5m, purges jobs older than 15m)
	api.StartJobCleaner(5*time.Minute, 15*time.Minute)

	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	http.HandleFunc("/info", api.GetInfo)
	http.HandleFunc("/download", api.StartDownload(sem))
	http.HandleFunc("/cancel/", api.CancelJob)
	http.HandleFunc("/progress/", api.GetProgress)
	http.HandleFunc("/file/", api.ServeFile)
	http.HandleFunc("/health", api.GetHealth)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server running on :" + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server stopped: %v\n", err)
	}
}
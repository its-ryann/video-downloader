package server

import (
	"fmt"
	"net/http"

	"video-downloader/internal/api"
)

func Start() {
	// Creates a semaphore with 3 slots - max of 3 downloads at once.
	sem := make(chan struct{}, 3)

	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	http.HandleFunc("/download", api.DownloadHandler(sem))

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
package server

import (
	"fmt"
	"net/http"

	"video-downloader/internal/api"
)

func Start() {
	http.HandleFunc("/download", api.DownloadHandler)

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
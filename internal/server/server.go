package server

import (
	"fmt"
	"net/http"
	"video-downloader/internal/api"
)

func Start() {
	sem := make(chan struct{}, 3)

	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	http.HandleFunc("/download", api.StartDownload(sem))
	http.HandleFunc("/progress/", api.GetProgress)
	http.HandleFunc("/file/", api.ServeFile)

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
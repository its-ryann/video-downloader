package server

import (
	"fmt"
	"os"
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server running on :" + port)
	http.ListenAndServe(":"+port, nil)
}
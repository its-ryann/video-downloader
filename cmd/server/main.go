package main

import (
	"fmt"
	"net/http"

	"video-downloader/internal/api"
)

func main() {
	// Root route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Server is running>>")
	})

	// Download endpoint
	http.HandleFunc("/download", api.DownloadHandler)

	fmt.Println("Starting server on :8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
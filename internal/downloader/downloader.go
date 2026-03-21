package downloader

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func DownloadVideo(url string) (string, error) {
	fileName := fmt.Sprintf("%d.mp4", time.Now().UnixNano())
	outPath := filepath.Join(os.TempDir(), fileName)

	cmd := exec.Command(
        "yt-dlp",
        "-f", "best",
        "--merge-output-format", "mp4",
        "-o", outPath,   // ✅ tell yt-dlp exactly where to write
        url,
    )

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to download video: %v\nOuput: %s", err, string(output))
	}

	// Find the downloaded file
	if _, err := os.Stat(outPath); err != nil {
        return "", fmt.Errorf("download completed but file not found: %v", err)
    }

	return outPath, nil
}
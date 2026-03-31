package downloader

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func DownloadVideo(url string, format string) (string, error) {
	// Default is mp4 if no format is specified
	if format == "" {
		format = "mp4"
	}

	// Use a unique temp file name to avoid conflicts
	outPath := filepath.Join(os.TempDir(), fmt.Sprintf("%d.%s", time.Now().UnixNano(), format))

	// Build yt-dlp arguments based on format
	args := []string{
		"--ignore-config",
		"--no-warnings",
		"-o", outPath,
	}

	if format == "mp3" {
		//Extract only the audio
		args = append(args, "-x", "--audio-format", "mp3")
	} else {
		args = append(args, "--merge-output-format", format)
	}

	args = append(args, url)

	cmd := exec.Command("yt-dlp", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp failed: %v\nOutput: %s", err, string(output))
	}

	if _, err := os.Stat(outPath); err != nil {
		return "", fmt.Errorf("file not found after download: %v", err)
	}

	return outPath, nil
}

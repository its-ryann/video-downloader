package downloader

import (
	"fmt"
	"os/exec"
)

func DownloadVideo(url string) (string, error) {
	outputFile := "video.%(ext)s"

	cmd := exec.Command("yt-dlp", "-f", "best", "-o", outputFile, url)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to download video: %v\nOuput: %s", err, string(output))
	}
	// Note: The actual output file name will depend on the video format and extension,
	return outputFile, nil
}
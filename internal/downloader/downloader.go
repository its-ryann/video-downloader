package downloader

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func DownloadVideo(url string, format string, onProgress func(int)) (string, error) {
	if format == "" {
		format = "mp4"
	}

	outPath := filepath.Join(os.TempDir(), fmt.Sprintf("%d.%s", time.Now().UnixNano(), format))

	args := []string{
		"--ignore-config",
		"--no-warnings",
		"--newline",
		"-o", outPath,
	}

	if format == "mp3" {
		args = append(args, "-x", "--audio-format", "mp3")
	} else {
		args = append(args, "--merge-output-format", format)
	}

	args = append(args, url)

	cmd := exec.Command("yt-dlp", args...)

	// Pipe stdout to read progress line by line
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to pipe stdout: %v", err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start yt-dlp: %v", err)
	}

	// Read each line yt-dlp prints and extract the percentage
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		// yt-dlp progress lines look like: "[download]  45.3% of ..."
		if strings.Contains(line, "[download]") && strings.Contains(line, "%") {
			fields := strings.Fields(line)
			for _, f := range fields {
				if strings.HasSuffix(f, "%") {
					pct, err := strconv.ParseFloat(strings.TrimSuffix(f, "%"), 64)
					if err == nil && onProgress != nil {
						onProgress(int(pct))
					}
					break
				}
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("yt-dlp failed: %v", err)
	}

	if _, err := os.Stat(outPath); err != nil {
		return "", fmt.Errorf("file not found after download: %v", err)
	}

	return outPath, nil
}
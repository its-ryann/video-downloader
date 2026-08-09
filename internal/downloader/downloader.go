package downloader

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type VideoInfo struct {
	Title     string  `json:"title"`
	Thumbnail string  `json:"thumbnail"`
	Duration  float64 `json:"duration"`
	Uploader  string  `json:"uploader"`
	Platform  string  `json:"platform"`
}

type ytdlpDump struct {
	Title     string  `json:"title"`
	Thumbnail string  `json:"thumbnail"`
	Duration  float64 `json:"duration"`
	Uploader  string  `json:"uploader"`
	Extractor string  `json:"extractor_key"`
}

func GetVideoInfo(ctx context.Context, videoURL string) (*VideoInfo, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctxTimeout, "yt-dlp", "--dump-json", "--no-warnings", "--skip-download", videoURL)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video metadata: %v", err)
	}

	var dump ytdlpDump
	if err := json.Unmarshal(out, &dump); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %v", err)
	}

	info := &VideoInfo{
		Title:     dump.Title,
		Thumbnail: dump.Thumbnail,
		Duration:  dump.Duration,
		Uploader:  dump.Uploader,
		Platform:  dump.Extractor,
	}
	return info, nil
}

func DownloadVideo(ctx context.Context, jobID string, videoURL string, format string, quality string, onProgress func(int)) (string, string, error) {
	if format == "" {
		format = "mp4"
	}

	outTemplate := filepath.Join(os.TempDir(), fmt.Sprintf("vidsnap_%s.%%(ext)s", jobID))

	args := []string{
		"--ignore-config",
		"--no-warnings",
		"--newline",
		"-o", outTemplate,
	}

	if format == "mp3" {
		args = append(args, "-x", "--audio-format", "mp3", "--audio-quality", "0")
	} else {
		var formatSelector string
		switch quality {
		case "1080p":
			formatSelector = "bestvideo[height<=1080]+bestaudio/best[height<=1080]/best"
		case "720p":
			formatSelector = "bestvideo[height<=720]+bestaudio/best[height<=720]/best"
		case "480p":
			formatSelector = "bestvideo[height<=480]+bestaudio/best[height<=480]/best"
		case "360p":
			formatSelector = "bestvideo[height<=360]+bestaudio/best[height<=360]/best"
		default:
			formatSelector = "bestvideo+bestaudio/best"
		}
		args = append(args, "-f", formatSelector, "--merge-output-format", format)
	}

	args = append(args, videoURL)

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to pipe stdout: %v", err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("failed to start yt-dlp: %v", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[download]") && strings.Contains(line, "%") {
			fields := strings.Fields(line)
			for _, f := range fields {
				if strings.HasSuffix(f, "%") {
					pctStr := strings.TrimSuffix(f, "%")
					pct, err := strconv.ParseFloat(pctStr, 64)
					if err == nil && onProgress != nil {
						onProgress(int(pct))
					}
					break
				}
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		// Cleanup any partial files on error / cancellation
		matches, _ := filepath.Glob(filepath.Join(os.TempDir(), fmt.Sprintf("vidsnap_%s*", jobID)))
		for _, m := range matches {
			os.Remove(m)
		}
		if ctx.Err() == context.Canceled {
			return "", "", fmt.Errorf("download cancelled")
		}
		return "", "", fmt.Errorf("yt-dlp failed: %v", err)
	}

	matches, err := filepath.Glob(filepath.Join(os.TempDir(), fmt.Sprintf("vidsnap_%s.*", jobID)))
	if err != nil || len(matches) == 0 {
		return "", "", fmt.Errorf("downloaded file not found")
	}

	actualFile := matches[0]
	ext := filepath.Ext(actualFile)

	return actualFile, ext, nil
}

func SanitizeFilename(name string) string {
	reg := regexp.MustCompile(`[^\w\s\-\.\(\)]+`)
	sanitized := reg.ReplaceAllString(name, "_")
	sanitized = strings.TrimSpace(sanitized)
	if sanitized == "" {
		return "vidsnap_download"
	}
	return sanitized
}
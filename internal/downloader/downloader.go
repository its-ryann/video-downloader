package downloader

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func DownloadVideo(url string) (string, error) {
	cmd := exec.Command(
		"yt-dlp",
		"--ignore-config",                // ignore any system/user config files
		"--no-warnings",                  // suppress non-critical warnings
		"--merge-output-format", "mp4",   // ensure output is mp4
		"--print", "after_move:filepath", // print final file path only
		url,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to pipe stdout: %v", err)
	}

	// Extract last non-empty line — the file path is always last
	filePath := extractFilePath(string(output))

	if filePath == "" {
		return "", fmt.Errorf("could not determine downloaded file path.\nRaw output: %s", string(output))
	}

	return filePath, nil
}

// extractFilePath gets the last non-empty line from yt-dlp output
// This is defensive — even if warnings sneak through, the file path is always last
func extractFilePath(output string) string {
	lines := strings.Split(output, "\n")

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		// File path will start with / (absolute path on Linux)
		if strings.HasPrefix(line, "/") {
			return line
		}
	}

	return ""
}
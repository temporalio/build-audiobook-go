// @@@SNIPSTART audiobook-project-go-tts-activities
package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.temporal.io/sdk/temporal"
)

type Activities struct {
	BearerToken string
	fileMutex   sync.Mutex
}

const (
	apiEndpoint          = "https://api.openai.com/v1/audio/speech"
	contentType          = "application/json"
	requestTimeout       = 30 * time.Second
	maxTokens            = 512
	averageTokensPerWord = 1.33
	fileExtension        = ".mp3"
)

func (a *Activities) ReadFile(ctx context.Context, inputPath string) ([]string, error) {
	if inputPath == "" || !strings.HasSuffix(inputPath, ".txt") {
		return nil, temporal.NewApplicationError("Invalid path", "MALFORMED_INPUT", nil)
	}

	if strings.HasPrefix(inputPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, temporal.NewApplicationError("Unable to determine home directory", "HOME_DIR_ERROR", err)
		}
		inputPath = filepath.Join(home, inputPath[1:])
	}

	canonicalPath, err := filepath.Abs(inputPath)
	if err != nil {
		return nil, temporal.NewApplicationError("Invalid path", "MALFORMED_INPUT", err)
	}

	fileInfo, err := os.Stat(canonicalPath)
	if err != nil || fileInfo.IsDir() || !fileInfo.Mode().IsRegular() {
		return nil, temporal.NewApplicationError("Invalid path", "MALFORMED_INPUT", err)
	}

	content, err := os.ReadFile(canonicalPath)
	if err != nil {
		return nil, temporal.NewApplicationError("Invalid content", "MISSING_CONTENT", err)
	}

	trimmedContent := strings.TrimSpace(string(content))
	words := strings.Fields(trimmedContent)
	chunks := []string{}
	chunk := strings.Builder{}

	for _, word := range words {
		if float64(chunk.Len()+len(word))*averageTokensPerWord <= maxTokens {
			if chunk.Len() > 0 {
				chunk.WriteString(" ")
			}
			chunk.WriteString(word)
		} else {
			chunks = append(chunks, chunk.String())
			chunk.Reset()
			chunk.WriteString(word)
		}
	}

	if chunk.Len() > 0 {
		chunks = append(chunks, chunk.String())
	}

	return chunks, nil
}

func (a *Activities) CreateTemporaryFile(ctx context.Context) (string, error) {
	tempFile, err := os.CreateTemp("", "*.tmp")
	if err != nil {
		return "", temporal.NewApplicationError("Unable to create temporary work file", "FILE_ERROR", err)
	}

	if err := tempFile.Close(); err != nil {
		return "", temporal.NewApplicationError("Unable to close temporary work file", "FILE_ERROR", err)
	}

	return tempFile.Name(), nil
}

func (a *Activities) Process(ctx context.Context, chunk, outputPath string) error {
	reqBody := fmt.Sprintf(`{
        "model": "tts-1",
        "input": %q,
        "voice": "nova",
        "response_format": "mp3"
    }`, chunk)

	client := &http.Client{
		Timeout: requestTimeout,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiEndpoint, strings.NewReader(reqBody))
	if err != nil {
		return temporal.NewApplicationError("Failed to create request", "REQUEST_ERROR", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.BearerToken)
	req.Header.Set("Content-Type", contentType)

	resp, err := client.Do(req)
	if err != nil {
		return temporal.NewApplicationError("Failed to execute request", "REQUEST_ERROR", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return temporal.NewApplicationError(fmt.Sprintf("Received Unexpected status code: %d", resp.StatusCode), "REQUEST_ERROR", nil)
	}

	if resp.Header.Get("Content-Type") != "audio/mpeg" {
		return temporal.NewApplicationError("Received unexpected content type", "RESPONSE_ERROR", nil)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return temporal.NewApplicationError("Failed to read response body", "RESPONSE_ERROR", err)
	}

	if len(body) == 0 {
		return temporal.NewApplicationError("Received empty response body", "RESPONSE_ERROR", nil)
	}

	file, err := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return temporal.NewApplicationError("Unable to open file for appending", "FILE_ERROR", err)
	}
	defer file.Close()

	_, err = file.Write(body)
	if err != nil {
		return temporal.NewApplicationError("Unable to write data to file", "FILE_ERROR", err)
	}

	return nil
}

func (a *Activities) MoveOutputFileToPlace(ctx context.Context, tempPath, originalPath string) (string, error) {
	baseName := strings.TrimSuffix(filepath.Base(originalPath), filepath.Ext(originalPath))
	parentDir := filepath.Dir(originalPath)
	newPath := filepath.Join(parentDir, baseName+fileExtension)

	a.fileMutex.Lock()
	defer a.fileMutex.Unlock()

	for i := 1; ; i++ {
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			break
		}
		newPath = filepath.Join(parentDir, fmt.Sprintf("%s-%d%s", baseName, i, fileExtension))
	}

	tempFile, err := os.Open(tempPath)
	if err != nil {
		return "", temporal.NewApplicationError("Unable to open temporary file", "FILE_ERROR", err)
	}
	defer tempFile.Close()

	fileInfo, err := tempFile.Stat()
	if err != nil {
		return "", temporal.NewApplicationError("Unable to get file info", "FILE_ERROR", err)
	}

	err = os.Rename(tempPath, newPath)
	if err != nil {
		return "", temporal.NewApplicationError("Unable to move output file to destination", "FILE_ERROR", err)
	}

	err = os.Chmod(newPath, fileInfo.Mode())
	if err != nil {
		return "", temporal.NewApplicationError("Unable to set file permissions", "FILE_ERROR", err)
	}

	return newPath, nil
}

// @@@SNIPEND

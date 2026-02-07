package history

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"
)

// Entry represents a single transcription history record.
type Entry struct {
	Timestamp        string  `json:"timestamp"`
	RawTranscription string  `json:"raw_transcription"`
	ProcessedText    string  `json:"processed_text"`
	AudioFile        string  `json:"audio_file"`
	DurationSec      float64 `json:"duration_sec"`
}

// historyDirFunc is overridable for testing.
var historyDirFunc = defaultHistoryDir

func defaultHistoryDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".voicecode", "history")
}

// HistoryDir returns the path to the history directory.
func HistoryDir() string {
	return historyDirFunc()
}

// Save writes WAV data and JSON metadata to the history directory.
// Returns the base filename (without extension).
func Save(wavData []byte, raw, processed string, durationSec float64) (string, error) {
	dir := historyDirFunc()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating history directory: %w", err)
	}

	baseName := time.Now().Format("2006-01-02_150405")
	wavPath := filepath.Join(dir, baseName+".wav")
	jsonPath := filepath.Join(dir, baseName+".json")

	if err := os.WriteFile(wavPath, wavData, 0o644); err != nil {
		return "", fmt.Errorf("writing WAV file: %w", err)
	}

	entry := Entry{
		Timestamp:        time.Now().Format(time.RFC3339),
		RawTranscription: raw,
		ProcessedText:    processed,
		AudioFile:        baseName + ".wav",
		DurationSec:      math.Round(durationSec*10) / 10,
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling entry: %w", err)
	}

	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		return "", fmt.Errorf("writing JSON file: %w", err)
	}

	return baseName, nil
}

package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withTempHistoryDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "history")
	historyDirFunc = func() string { return dir }
	t.Cleanup(func() { historyDirFunc = defaultHistoryDir })
	return dir
}

func TestSaveCreatesFiles(t *testing.T) {
	dir := withTempHistoryDir(t)

	wavData := []byte("fake wav data for testing")
	raw := "Hello World"
	processed := "hello_world"
	duration := 2.5

	baseName, err := Save(wavData, raw, processed, duration)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if baseName == "" {
		t.Fatal("Save() returned empty baseName")
	}

	// Verify WAV file exists and has correct content
	wavPath := filepath.Join(dir, baseName+".wav")
	wavContent, err := os.ReadFile(wavPath)
	if err != nil {
		t.Fatalf("WAV file not created: %v", err)
	}
	if string(wavContent) != string(wavData) {
		t.Errorf("WAV content mismatch")
	}

	// Verify JSON file exists
	jsonPath := filepath.Join(dir, baseName+".json")
	jsonContent, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("JSON file not created: %v", err)
	}

	var entry Entry
	if err := json.Unmarshal(jsonContent, &entry); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}

	if entry.RawTranscription != raw {
		t.Errorf("RawTranscription = %q, want %q", entry.RawTranscription, raw)
	}
	if entry.ProcessedText != processed {
		t.Errorf("ProcessedText = %q, want %q", entry.ProcessedText, processed)
	}
	if entry.DurationSec != duration {
		t.Errorf("DurationSec = %v, want %v", entry.DurationSec, duration)
	}
	if entry.AudioFile != baseName+".wav" {
		t.Errorf("AudioFile = %q, want %q", entry.AudioFile, baseName+".wav")
	}
}

func TestSaveFilenameFormat(t *testing.T) {
	withTempHistoryDir(t)

	baseName, err := Save([]byte("test"), "raw", "processed", 1.0)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Filename should match YYYY-MM-DD_HHMMSS format
	parts := strings.Split(baseName, "_")
	if len(parts) != 2 {
		t.Errorf("baseName %q does not match expected format YYYY-MM-DD_HHMMSS", baseName)
	}

	datePart := parts[0]
	if len(datePart) != 10 {
		t.Errorf("date part %q does not match YYYY-MM-DD", datePart)
	}

	timePart := parts[1]
	if len(timePart) != 6 {
		t.Errorf("time part %q does not match HHMMSS", timePart)
	}
}

func TestSaveCreatesHistoryDir(t *testing.T) {
	dir := withTempHistoryDir(t)

	// Dir should not exist yet
	if _, err := os.Stat(dir); err == nil {
		t.Fatal("history dir should not exist before Save")
	}

	_, err := Save([]byte("test"), "raw", "processed", 1.0)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if _, err := os.Stat(dir); err != nil {
		t.Errorf("history dir was not created: %v", err)
	}
}

func TestSaveDurationRounding(t *testing.T) {
	withTempHistoryDir(t)

	// 2.368 should round to 2.4
	baseName, err := Save([]byte("test"), "raw", "processed", 2.368)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	dir := historyDirFunc()
	jsonContent, err := os.ReadFile(filepath.Join(dir, baseName+".json"))
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	var entry Entry
	if err := json.Unmarshal(jsonContent, &entry); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}

	if entry.DurationSec != 2.4 {
		t.Errorf("DurationSec = %v, want 2.4", entry.DurationSec)
	}
}

func TestHistoryDir(t *testing.T) {
	dir := HistoryDir()
	if dir == "" {
		t.Error("HistoryDir() returned empty string")
	}
	if !strings.Contains(dir, ".voicecode") {
		t.Errorf("HistoryDir() = %q, expected to contain .voicecode", dir)
	}
}

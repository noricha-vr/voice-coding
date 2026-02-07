//go:build integration

package transcriber

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// sampleWAVPath returns the path to testdata/sample.wav relative to this test file.
func sampleWAVPath(t *testing.T) string {
	t.Helper()
	path := filepath.Join("..", "..", "..", "testdata", "sample.wav")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("testdata/sample.wav not found: %s", path)
	}
	return path
}

func TestTranscribeIntegration(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	ctx := context.Background()
	tr, err := New(ctx, apiKey)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	wavPath := sampleWAVPath(t)

	text, elapsed, err := tr.Transcribe(ctx, wavPath)
	if err != nil {
		t.Fatalf("Transcribe: %v", err)
	}

	t.Logf("Model: %s", tr.ModelName())
	t.Logf("Elapsed: %.2fs", elapsed)
	t.Logf("Text: %s", text)

	if text == "" {
		t.Log("Warning: transcription returned empty text (may be silence/hallucination)")
	}
}

func TestTranscribeMediumWAV(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	ctx := context.Background()
	tr, err := New(ctx, apiKey)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	wavPath := sampleWAVPath(t)

	text, elapsed, err := tr.Transcribe(ctx, wavPath)
	if err != nil {
		t.Fatalf("Transcribe: %v", err)
	}

	t.Logf("Model: %s", tr.ModelName())
	t.Logf("Elapsed: %.2fs", elapsed)
	t.Logf("Text: %s", text)
}

//go:build integration

package transcriber

import (
	"context"
	"os"
	"testing"
)

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

	// Use a small test WAV
	wavPath := os.Getenv("HOME") + "/.voicecode/history/2026-01-10_203724.wav"
	if _, err := os.Stat(wavPath); os.IsNotExist(err) {
		t.Skipf("test WAV not found: %s", wavPath)
	}

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

	wavPath := os.Getenv("HOME") + "/.voicecode/history/2026-01-09_202239.wav"
	if _, err := os.Stat(wavPath); os.IsNotExist(err) {
		t.Skipf("test WAV not found: %s", wavPath)
	}

	text, elapsed, err := tr.Transcribe(ctx, wavPath)
	if err != nil {
		t.Fatalf("Transcribe: %v", err)
	}

	t.Logf("Model: %s", tr.ModelName())
	t.Logf("Elapsed: %.2fs", elapsed)
	t.Logf("Text: %s", text)
}

package audio

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestGetDurationSampleWAV(t *testing.T) {
	path := filepath.Join("..", "..", "..", "testdata", "sample.wav")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("testdata/sample.wav not found: %v", err)
	}

	duration, err := GetDuration(path)
	if err != nil {
		t.Fatalf("GetDuration() error: %v", err)
	}

	// sample.wav is 75820 bytes, 16-bit mono 16kHz
	// data size = 75820 - 44 (header) = 75776 bytes
	// samples = 75776 / 2 = 37888
	// duration = 37888 / 16000 = 2.368 -> 2.4 seconds
	if duration <= 0 || duration > 60 {
		t.Errorf("GetDuration() = %v, expected a reasonable duration (0-60s)", duration)
	}

	t.Logf("sample.wav duration: %.1f seconds", duration)
}

func TestWriteWAVAndGetDurationRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.wav")

	sampleRate := 16000
	durationSec := 2.0
	numSamples := int(durationSec * float64(sampleRate))
	samples := make([]int16, numSamples)
	// Fill with a simple sine wave
	for i := range samples {
		samples[i] = int16(1000 * math.Sin(2*math.Pi*440*float64(i)/float64(sampleRate)))
	}

	if err := WriteWAV(path, samples, sampleRate); err != nil {
		t.Fatalf("WriteWAV() error: %v", err)
	}

	// Verify file was created
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("file is empty")
	}

	// Read back and check duration
	got, err := GetDuration(path)
	if err != nil {
		t.Fatalf("GetDuration() error: %v", err)
	}

	if math.Abs(got-durationSec) > 0.1 {
		t.Errorf("GetDuration() = %.1f, want %.1f", got, durationSec)
	}
}

func TestGetDurationInvalidFile(t *testing.T) {
	dir := t.TempDir()

	// Non-existent file
	_, err := GetDuration(filepath.Join(dir, "nonexistent.wav"))
	if err == nil {
		t.Error("GetDuration() should fail for non-existent file")
	}

	// Not a WAV file
	notWav := filepath.Join(dir, "not.wav")
	os.WriteFile(notWav, []byte("this is not a wav file"), 0o644)
	_, err = GetDuration(notWav)
	if err == nil {
		t.Error("GetDuration() should fail for non-WAV file")
	}
}

func TestWriteWAVEmptySamples(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.wav")

	if err := WriteWAV(path, []int16{}, 16000); err != nil {
		t.Fatalf("WriteWAV() error: %v", err)
	}

	got, err := GetDuration(path)
	if err != nil {
		t.Fatalf("GetDuration() error: %v", err)
	}
	if got != 0 {
		t.Errorf("GetDuration() = %.1f, want 0", got)
	}
}

package audio

import (
	"testing"
)

func TestTrimSilence_AllZero(t *testing.T) {
	samples := make([]int16, 16000) // 1s of silence
	trimmed, info := TrimSilence(samples, 16000)
	if !info.AllSilence {
		t.Fatalf("expected AllSilence=true, got false (info=%+v)", info)
	}
	if trimmed != nil && len(trimmed) != 0 {
		t.Fatalf("expected empty trimmed samples, got len=%d", len(trimmed))
	}
}

func TestTrimSilence_TrimsLeadingAndTrailing(t *testing.T) {
	const sr = 16000

	// 1s silence + 1s speech + 1s silence
	silence1 := make([]int16, sr)
	speech := make([]int16, sr)
	for i := range speech {
		speech[i] = 2000
	}
	silence2 := make([]int16, sr)

	samples := append(append(silence1, speech...), silence2...)

	trimmed, info := TrimSilence(samples, sr)
	if info.AllSilence {
		t.Fatalf("expected AllSilence=false, got true (info=%+v)", info)
	}
	if len(trimmed) >= len(samples) {
		t.Fatalf("expected trimmed to be smaller than original: trimmed=%d original=%d info=%+v", len(trimmed), len(samples), info)
	}

	// With 200ms pad on both sides, expected duration ~1.4s.
	wantMin := int(1.2 * float64(sr))
	wantMax := int(1.6 * float64(sr))
	if len(trimmed) < wantMin || len(trimmed) > wantMax {
		t.Fatalf("unexpected trimmed length: got=%d want=[%d,%d] info=%+v", len(trimmed), wantMin, wantMax, info)
	}
	if info.LeadingTrimmedSamples <= 0 || info.TrailingTrimmedSamples <= 0 {
		t.Fatalf("expected both leading and trailing trim > 0, got lead=%d tail=%d", info.LeadingTrimmedSamples, info.TrailingTrimmedSamples)
	}
}

func TestTrimSilence_SpeechAtStartDoesNotUnderflow(t *testing.T) {
	const sr = 16000

	// 1s speech + 1s silence
	speech := make([]int16, sr)
	for i := range speech {
		speech[i] = 2000
	}
	silence := make([]int16, sr)
	samples := append(speech, silence...)

	trimmed, info := TrimSilence(samples, sr)
	if info.AllSilence {
		t.Fatalf("expected AllSilence=false, got true (info=%+v)", info)
	}
	// Start should be clamped to 0, so no leading trim.
	if info.LeadingTrimmedSamples != 0 {
		t.Fatalf("expected LeadingTrimmedSamples=0, got %d info=%+v", info.LeadingTrimmedSamples, info)
	}
	if len(trimmed) == 0 {
		t.Fatalf("expected non-empty trimmed output")
	}
}

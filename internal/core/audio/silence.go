package audio

import (
	"math"
	"sort"
)

type TrimInfo struct {
	OriginalSamples        int
	TrimmedSamples         int
	LeadingTrimmedSamples  int
	TrailingTrimmedSamples int
	WindowSamples          int
	PadSamples             int
	NoiseFloor             float64
	Threshold              float64
	AllSilence             bool
}

// TrimSilence trims leading and trailing silence from PCM samples.
//
// Heuristic:
// - Compute mean absolute amplitude per 10ms window.
// - Estimate noise floor via 10th percentile of window energies.
// - Detect first/last window above threshold, then pad both sides (200ms).
//
// It returns the trimmed samples and metadata for logging/diagnostics.
func TrimSilence(samples []int16, sampleRate int) ([]int16, TrimInfo) {
	info := TrimInfo{OriginalSamples: len(samples)}
	if len(samples) == 0 || sampleRate <= 0 {
		info.AllSilence = len(samples) == 0
		return samples, info
	}

	// 10ms windows
	windowSamples := sampleRate / 100
	if windowSamples < 1 {
		windowSamples = 1
	}
	info.WindowSamples = windowSamples

	numWindows := (len(samples) + windowSamples - 1) / windowSamples
	energies := make([]float64, 0, numWindows)

	var maxEnergy float64
	for i := 0; i < len(samples); i += windowSamples {
		end := i + windowSamples
		if end > len(samples) {
			end = len(samples)
		}
		var sum float64
		for _, s := range samples[i:end] {
			sum += math.Abs(float64(s))
		}
		mean := sum / float64(end-i)
		energies = append(energies, mean)
		if mean > maxEnergy {
			maxEnergy = mean
		}
	}

	if maxEnergy == 0 {
		info.AllSilence = true
		return nil, info
	}

	noiseFloor := percentile(energies, 0.10)
	info.NoiseFloor = noiseFloor

	// Conservative threshold: depends on both noise floor and max energy.
	threshold := math.Max(noiseFloor*3.0, maxEnergy*0.05)
	threshold = math.Max(threshold, 80.0)
	info.Threshold = threshold

	startW := -1
	for i, e := range energies {
		if e >= threshold {
			startW = i
			break
		}
	}
	endW := -1
	for i := len(energies) - 1; i >= 0; i-- {
		if energies[i] >= threshold {
			endW = i
			break
		}
	}

	if startW == -1 || endW == -1 || endW < startW {
		// Treat as silence only when max energy is extremely small.
		// Otherwise, keep the original to avoid false negatives for quiet speech.
		if maxEnergy < 80.0 {
			info.AllSilence = true
			return nil, info
		}
		info.TrimmedSamples = len(samples)
		return samples, info
	}

	padSamples := int(float64(sampleRate) * 0.20) // 200ms
	if padSamples < 0 {
		padSamples = 0
	}
	info.PadSamples = padSamples

	start := startW*windowSamples - padSamples
	if start < 0 {
		start = 0
	}
	end := (endW+1)*windowSamples + padSamples
	if end > len(samples) {
		end = len(samples)
	}
	if end <= start {
		info.AllSilence = true
		return nil, info
	}

	info.TrimmedSamples = end - start
	info.LeadingTrimmedSamples = start
	info.TrailingTrimmedSamples = len(samples) - end

	trimmed := samples[start:end]
	return trimmed, info
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if p <= 0 {
		min := values[0]
		for _, v := range values[1:] {
			if v < min {
				min = v
			}
		}
		return min
	}
	if p >= 1 {
		max := values[0]
		for _, v := range values[1:] {
			if v > max {
				max = v
			}
		}
		return max
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	// Index in [0, len-1]
	pos := p * float64(len(sorted)-1)
	idx := int(math.Round(pos))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

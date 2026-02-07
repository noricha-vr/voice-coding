package audio

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

const (
	riffHeaderSize = 12
	fmtChunkSize   = 24
	dataChunkID    = "data"
	pcmFormat      = 1
	bitsPerSample  = 16
)

// GetDuration reads a WAV file and returns its duration in seconds,
// rounded to 1 decimal place.
func GetDuration(path string) (float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("opening WAV file: %w", err)
	}
	defer f.Close()

	// Read RIFF header
	var riffID [4]byte
	if err := binary.Read(f, binary.LittleEndian, &riffID); err != nil {
		return 0, fmt.Errorf("reading RIFF ID: %w", err)
	}
	if string(riffID[:]) != "RIFF" {
		return 0, fmt.Errorf("not a RIFF file")
	}

	var fileSize uint32
	if err := binary.Read(f, binary.LittleEndian, &fileSize); err != nil {
		return 0, fmt.Errorf("reading file size: %w", err)
	}

	var waveID [4]byte
	if err := binary.Read(f, binary.LittleEndian, &waveID); err != nil {
		return 0, fmt.Errorf("reading WAVE ID: %w", err)
	}
	if string(waveID[:]) != "WAVE" {
		return 0, fmt.Errorf("not a WAVE file")
	}

	// Parse chunks
	var channels, sampleRate, bps uint16
	var dataSize uint32
	foundFmt := false
	foundData := false

	for {
		var chunkID [4]byte
		if err := binary.Read(f, binary.LittleEndian, &chunkID); err != nil {
			break
		}

		var chunkSize uint32
		if err := binary.Read(f, binary.LittleEndian, &chunkSize); err != nil {
			break
		}

		id := string(chunkID[:])

		switch id {
		case "fmt ":
			var audioFormat uint16
			if err := binary.Read(f, binary.LittleEndian, &audioFormat); err != nil {
				return 0, fmt.Errorf("reading audio format: %w", err)
			}
			if err := binary.Read(f, binary.LittleEndian, &channels); err != nil {
				return 0, fmt.Errorf("reading channels: %w", err)
			}
			var sr uint32
			if err := binary.Read(f, binary.LittleEndian, &sr); err != nil {
				return 0, fmt.Errorf("reading sample rate: %w", err)
			}
			sampleRate = uint16(sr)
			// Skip byte rate (4 bytes) and block align (2 bytes)
			var byteRate uint32
			var blockAlign uint16
			binary.Read(f, binary.LittleEndian, &byteRate)
			binary.Read(f, binary.LittleEndian, &blockAlign)
			if err := binary.Read(f, binary.LittleEndian, &bps); err != nil {
				return 0, fmt.Errorf("reading bits per sample: %w", err)
			}
			// Skip remaining fmt chunk data if any
			remaining := int64(chunkSize) - 16
			if remaining > 0 {
				f.Seek(remaining, 1)
			}
			foundFmt = true

		case "data":
			dataSize = chunkSize
			foundData = true
			// Skip past data to look for more chunks (not needed, we have what we need)
			break

		default:
			// Skip unknown chunk
			f.Seek(int64(chunkSize), 1)
		}

		if foundFmt && foundData {
			break
		}
	}

	if !foundFmt {
		return 0, fmt.Errorf("fmt chunk not found")
	}
	if !foundData {
		return 0, fmt.Errorf("data chunk not found")
	}
	if sampleRate == 0 || channels == 0 || bps == 0 {
		return 0, fmt.Errorf("invalid WAV header values")
	}

	bytesPerSample := uint32(channels) * uint32(bps) / 8
	totalSamples := dataSize / bytesPerSample
	duration := float64(totalSamples) / float64(sampleRate)
	duration = math.Round(duration*10) / 10

	return duration, nil
}

// WriteWAV writes PCM 16-bit mono audio data to a WAV file.
func WriteWAV(path string, samples []int16, sampleRate int) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating WAV file: %w", err)
	}
	defer f.Close()

	numChannels := uint16(1)
	bps := uint16(bitsPerSample)
	dataSize := uint32(len(samples)) * uint32(bps/8) * uint32(numChannels)
	// RIFF file size = 4 (WAVE) + 24 (fmt chunk) + 8 (data header) + dataSize
	riffSize := uint32(4 + fmtChunkSize + 8 + dataSize)

	// RIFF header
	f.Write([]byte("RIFF"))
	binary.Write(f, binary.LittleEndian, riffSize)
	f.Write([]byte("WAVE"))

	// fmt chunk
	f.Write([]byte("fmt "))
	binary.Write(f, binary.LittleEndian, uint32(16)) // chunk size
	binary.Write(f, binary.LittleEndian, uint16(pcmFormat))
	binary.Write(f, binary.LittleEndian, numChannels)
	binary.Write(f, binary.LittleEndian, uint32(sampleRate))
	byteRate := uint32(sampleRate) * uint32(numChannels) * uint32(bps/8)
	binary.Write(f, binary.LittleEndian, byteRate)
	blockAlign := numChannels * (bps / 8)
	binary.Write(f, binary.LittleEndian, blockAlign)
	binary.Write(f, binary.LittleEndian, bps)

	// data chunk
	f.Write([]byte("data"))
	binary.Write(f, binary.LittleEndian, dataSize)
	if err := binary.Write(f, binary.LittleEndian, samples); err != nil {
		return fmt.Errorf("writing samples: %w", err)
	}

	return nil
}

//go:build darwin

package recorder

import (
	"fmt"
	"log"
	"sync"

	"github.com/gordonklaus/portaudio"
)

const (
	SampleRate = 16000
	Channels   = 1
	FrameSize  = 1024
)

type darwinRecorder struct {
	mu        sync.Mutex
	stream    *portaudio.Stream
	buffer    []int16
	recording bool
}

// NewRecorder creates a new macOS audio recorder using PortAudio.
func NewRecorder() (Recorder, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("portaudio init: %w", err)
	}
	return &darwinRecorder{}, nil
}

func (r *darwinRecorder) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.recording {
		return fmt.Errorf("already recording")
	}
	r.buffer = make([]int16, 0, SampleRate*120) // pre-alloc for 120s

	buf := make([]int16, FrameSize)
	stream, err := portaudio.OpenDefaultStream(Channels, 0, float64(SampleRate), FrameSize, buf)
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}
	if err := stream.Start(); err != nil {
		return fmt.Errorf("start stream: %w", err)
	}
	r.stream = stream
	r.recording = true

	go func() {
		for {
			r.mu.Lock()
			if !r.recording {
				r.mu.Unlock()
				return
			}
			r.mu.Unlock()

			if err := stream.Read(); err != nil {
				log.Printf("[Recorder] stream read error: %v", err)
				return
			}
			r.mu.Lock()
			r.buffer = append(r.buffer, buf...)
			r.mu.Unlock()
		}
	}()

	return nil
}

func (r *darwinRecorder) Stop() ([]int16, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.recording {
		return nil, fmt.Errorf("not recording")
	}
	r.recording = false
	if r.stream != nil {
		r.stream.Stop()
		r.stream.Close()
		r.stream = nil
	}
	samples := make([]int16, len(r.buffer))
	copy(samples, r.buffer)
	r.buffer = nil
	return samples, nil
}

func (r *darwinRecorder) IsRecording() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.recording
}

package recorder

import (
	"fmt"
	"sync"

	"github.com/gordonklaus/portaudio"
)

const (
	SampleRate = 16000
	Channels   = 1
	FrameSize  = 1024
)

type portaudioRecorder struct {
	mu        sync.Mutex
	stream    *portaudio.Stream
	buffer    []int16
	recording bool
}

// NewRecorder creates a new audio recorder using PortAudio.
func NewRecorder() (Recorder, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("portaudio init: %w", err)
	}
	return &portaudioRecorder{}, nil
}

func (r *portaudioRecorder) Start() error {
	r.mu.Lock()
	if r.recording {
		r.mu.Unlock()
		return fmt.Errorf("already recording")
	}

	// Reuse buffer capacity across recordings to avoid allocations in the PortAudio callback.
	// Prealloc 10s and let it grow if the user records longer.
	const preallocSeconds = 10
	preallocCap := SampleRate * preallocSeconds
	if cap(r.buffer) < preallocCap {
		r.buffer = make([]int16, 0, preallocCap)
	} else {
		r.buffer = r.buffer[:0]
	}
	r.recording = true
	r.mu.Unlock()

	// Use a callback stream to avoid the blocking Read() hang observed on Stop().
	// The callback appends input samples while recording=true.
	stream, err := portaudio.OpenDefaultStream(Channels, 0, float64(SampleRate), FrameSize, r.processAudio)
	if err != nil {
		r.mu.Lock()
		r.recording = false
		r.mu.Unlock()
		return fmt.Errorf("open stream: %w", err)
	}
	if err := stream.Start(); err != nil {
		_ = stream.Close()
		r.mu.Lock()
		r.recording = false
		r.mu.Unlock()
		return fmt.Errorf("start stream: %w", err)
	}

	r.mu.Lock()
	r.stream = stream
	r.mu.Unlock()

	return nil
}

func (r *portaudioRecorder) Stop() ([]int16, error) {
	r.mu.Lock()
	if !r.recording {
		r.mu.Unlock()
		return nil, fmt.Errorf("not recording")
	}
	r.recording = false
	stream := r.stream
	r.stream = nil
	r.mu.Unlock()

	var streamErr error
	if stream != nil {
		// Abort is best-effort and should be quick for callback streams.
		if err := stream.Abort(); err != nil {
			streamErr = fmt.Errorf("abort stream: %w", err)
		}
		if err := stream.Close(); err != nil {
			if streamErr != nil {
				streamErr = fmt.Errorf("%v; close stream: %w", streamErr, err)
			} else {
				streamErr = fmt.Errorf("close stream: %w", err)
			}
		}
	}

	r.mu.Lock()
	samples := make([]int16, len(r.buffer))
	copy(samples, r.buffer)
	// Keep capacity for the next recording to reduce allocations.
	r.buffer = r.buffer[:0]
	r.mu.Unlock()

	return samples, streamErr
}

func (r *portaudioRecorder) IsRecording() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.recording
}

func (r *portaudioRecorder) processAudio(in []int16) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.recording {
		return
	}
	r.buffer = append(r.buffer, in...)
}

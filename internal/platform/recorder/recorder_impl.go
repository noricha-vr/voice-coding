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

type portaudioRecorder struct {
	mu        sync.Mutex
	stream    *portaudio.Stream
	buffer    []int16
	frameBuf  []int16
	recording bool
	readDone  chan struct{}
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
	defer r.mu.Unlock()
	if r.recording {
		return fmt.Errorf("already recording")
	}

	// Reuse buffer capacity across recordings to avoid hot path allocations.
	// Prealloc 10s and let it grow if the user records longer.
	const preallocSeconds = 10
	preallocCap := SampleRate * preallocSeconds
	if cap(r.buffer) < preallocCap {
		r.buffer = make([]int16, 0, preallocCap)
	} else {
		r.buffer = r.buffer[:0]
	}
	if cap(r.frameBuf) < FrameSize {
		r.frameBuf = make([]int16, FrameSize)
	} else {
		r.frameBuf = r.frameBuf[:FrameSize]
	}

	buf := r.frameBuf
	stream, err := portaudio.OpenDefaultStream(Channels, 0, float64(SampleRate), FrameSize, buf)
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}
	if err := stream.Start(); err != nil {
		_ = stream.Close()
		return fmt.Errorf("start stream: %w", err)
	}
	r.stream = stream
	r.recording = true
	r.readDone = make(chan struct{})

	go func() {
		defer close(r.readDone)
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

func (r *portaudioRecorder) Stop() ([]int16, error) {
	r.mu.Lock()
	if !r.recording {
		r.mu.Unlock()
		return nil, fmt.Errorf("not recording")
	}
	r.recording = false
	stream := r.stream
	r.stream = nil
	done := r.readDone
	r.readDone = nil
	r.mu.Unlock()

	var streamErr error
	if stream != nil {
		if err := stream.Stop(); err != nil {
			streamErr = fmt.Errorf("stop stream: %w", err)
		}
		if err := stream.Close(); err != nil {
			if streamErr != nil {
				streamErr = fmt.Errorf("%v; close stream: %w", streamErr, err)
			} else {
				streamErr = fmt.Errorf("close stream: %w", err)
			}
		}
	}
	if done != nil {
		<-done
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

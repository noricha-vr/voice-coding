package app

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/noricha-vr/voicecode/internal/platform/sound"
)

type countingSound struct {
	mu     sync.Mutex
	counts map[sound.SoundType]int
}

func (c *countingSound) Play(s sound.SoundType) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.counts == nil {
		c.counts = make(map[sound.SoundType]int)
	}
	c.counts[s]++
	return nil
}

func (c *countingSound) Count(s sound.SoundType) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[s]
}

func TestStartPeriodicSoundPlaysUntilStopped(t *testing.T) {
	snd := &countingSound{}

	stop := startPeriodicSound(context.Background(), snd, sound.ProcessingTick, 10*time.Millisecond)

	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if snd.Count(sound.ProcessingTick) >= 2 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	stop()
	stop() // idempotent

	got := snd.Count(sound.ProcessingTick)
	if got < 2 {
		t.Fatalf("expected at least 2 ticks, got %d", got)
	}

	// After stop returns, the goroutine should be terminated and no more ticks should happen.
	time.Sleep(30 * time.Millisecond)
	after := snd.Count(sound.ProcessingTick)
	if after != got {
		t.Fatalf("tick count increased after stop: before=%d after=%d", got, after)
	}
}

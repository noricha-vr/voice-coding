package app

import (
	"context"
	"sync"
	"time"

	"github.com/noricha-vr/voicecode/internal/platform/sound"
)

// startPeriodicSound plays the given sound at the specified interval until stopped.
// It returns an idempotent stop func that blocks until the goroutine terminates.
func startPeriodicSound(ctx context.Context, player sound.Player, s sound.SoundType, interval time.Duration) (stop func()) {
	if player == nil {
		return func() {}
	}
	if interval <= 0 {
		interval = 1 * time.Second
	}

	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		defer close(done)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = player.Play(s)
			}
		}
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			cancel()
			<-done
		})
	}
}

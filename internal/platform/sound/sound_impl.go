package sound

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/noricha-vr/voicecode/assets"
)

const (
	wavHeaderSize = 44
)

var (
	otoCtx     *oto.Context
	otoOnce    sync.Once
	otoInitErr error
)

func initOto() {
	otoOnce.Do(func() {
		op := &oto.NewContextOptions{
			SampleRate:   22050,
			ChannelCount: 1,
			Format:       oto.FormatSignedInt16LE,
		}
		var readyCh chan struct{}
		otoCtx, readyCh, otoInitErr = oto.NewContext(op)
		if otoInitErr == nil {
			<-readyCh
		}
	})
}

var soundData = map[SoundType][]byte{
	Start:          nil,
	Stop:           nil,
	Success:        nil,
	Error:          nil,
	ProcessingTick: nil,
}

func init() {
	soundData[Start] = assets.SoundStart
	soundData[Stop] = assets.SoundStop
	soundData[Success] = assets.SoundSuccess
	soundData[Error] = assets.SoundError
	soundData[ProcessingTick] = assets.SoundProcessingTick
}

type otoPlayer struct{}

// NewPlayer creates a new cross-platform sound player.
func NewPlayer() Player { return &otoPlayer{} }

// WarmUp initializes the underlying audio backend so the first Play() call
// doesn't pay the initialization cost on the hot path.
func WarmUp() error {
	initOto()
	return otoInitErr
}

func (p *otoPlayer) Play(s SoundType) error {
	data, ok := soundData[s]
	if !ok {
		return fmt.Errorf("unknown sound type: %d", s)
	}

	initOto()
	if otoInitErr != nil {
		return fmt.Errorf("oto init: %w", otoInitErr)
	}

	if len(data) <= wavHeaderSize {
		return fmt.Errorf("invalid WAV data")
	}
	pcm := data[wavHeaderSize:]

	player := otoCtx.NewPlayer(bytes.NewReader(pcm))
	player.Play()
	go func() {
		for player.IsPlaying() {
			time.Sleep(10 * time.Millisecond)
		}
		player.Close()
	}()

	return nil
}

package app

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/noricha-vr/voicecode/internal/core/audio"
	"github.com/noricha-vr/voicecode/internal/core/history"
	"github.com/noricha-vr/voicecode/internal/core/settings"
	"github.com/noricha-vr/voicecode/internal/core/transcriber"
	"github.com/noricha-vr/voicecode/internal/platform/clipboard"
	"github.com/noricha-vr/voicecode/internal/platform/hotkey"
	"github.com/noricha-vr/voicecode/internal/platform/overlay"
	"github.com/noricha-vr/voicecode/internal/platform/recorder"
	"github.com/noricha-vr/voicecode/internal/platform/sound"
	"github.com/noricha-vr/voicecode/internal/platform/tray"
)

// App is the main application orchestrator.
type App struct {
	settings    *settings.Settings
	transcriber *transcriber.Transcriber
	recorder    recorder.Recorder
	clipboard   clipboard.Clipboard
	sound       sound.Player
	overlay     overlay.Overlay
	hotkey      hotkey.Manager
	tray        tray.Manager

	mu          sync.Mutex
	isRecording bool
	cancelTimer context.CancelFunc

	processMu sync.Mutex // guards processRecording from concurrent execution
}

// New creates a new App with all dependencies.
func New(
	cfg *settings.Settings,
	tr *transcriber.Transcriber,
	rec recorder.Recorder,
	clip clipboard.Clipboard,
	snd sound.Player,
	ov overlay.Overlay,
	hk hotkey.Manager,
	tm tray.Manager,
) *App {
	return &App{
		settings:    cfg,
		transcriber: tr,
		recorder:    rec,
		clipboard:   clip,
		sound:       snd,
		overlay:     ov,
		hotkey:      hk,
		tray:        tm,
	}
}

// Run starts the application with the tray icon.
func (a *App) Run() {
	a.tray.Run(func() {
		// onReady
		log.Println("[App] Ready")
		a.tray.SetState(tray.Idle)

		if a.settings.PushToTalk {
			a.hotkey.Register(a.settings.Hotkey, a.onHotkeyPress, a.onHotkeyRelease)
		} else {
			a.hotkey.Register(a.settings.Hotkey, a.onHotkeyPress, nil)
		}
		log.Printf("[App] Hotkey registered: %s (push-to-talk: %v)", a.settings.Hotkey, a.settings.PushToTalk)
	}, func() {
		// onQuit
		log.Println("[App] Shutting down")
		a.hotkey.Unregister()
		a.mu.Lock()
		recording := a.isRecording
		a.mu.Unlock()
		if recording {
			a.recorder.Stop()
		}
	})
}

func (a *App) onHotkeyPress() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isRecording {
		a.stopAndProcess()
	} else {
		a.startRecording()
	}
}

func (a *App) onHotkeyRelease() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isRecording && a.settings.PushToTalk {
		a.stopAndProcess()
	}
}

func (a *App) startRecording() {
	if err := a.recorder.Start(); err != nil {
		log.Printf("[App] Failed to start recording: %v", err)
		a.sound.Play(sound.Error)
		return
	}

	a.isRecording = true
	a.sound.Play(sound.Start)
	a.overlay.Show("Recording...")
	a.tray.SetState(tray.Recording)

	// Start timeout timer
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelTimer = cancel
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(a.settings.MaxRecordingDuration) * time.Second):
			log.Printf("[App] Recording timeout (%ds)", a.settings.MaxRecordingDuration)
			a.mu.Lock()
			defer a.mu.Unlock()
			if a.isRecording {
				a.stopAndProcess()
			}
		}
	}()
}

func (a *App) stopAndProcess() {
	if a.cancelTimer != nil {
		a.cancelTimer()
		a.cancelTimer = nil
	}

	samples, err := a.recorder.Stop()
	a.isRecording = false
	a.sound.Play(sound.Stop)
	a.overlay.Hide()
	a.tray.SetState(tray.Processing)

	if err != nil {
		log.Printf("[App] Failed to stop recording: %v", err)
		a.sound.Play(sound.Error)
		a.tray.SetState(tray.Idle)
		return
	}

	// Process in background to avoid blocking hotkey
	go a.processRecording(samples)
}

func (a *App) processRecording(samples []int16) {
	a.processMu.Lock()
	defer a.processMu.Unlock()
	a.tray.SetState(tray.Processing)
	defer a.tray.SetState(tray.Idle)

	// Save WAV to temp file
	tmpDir := os.TempDir()
	wavPath := filepath.Join(tmpDir, "voicecode_recording.wav")
	if err := audio.WriteWAV(wavPath, samples, 16000); err != nil {
		log.Printf("[App] Failed to write WAV: %v", err)
		a.sound.Play(sound.Error)
		return
	}
	defer os.Remove(wavPath)

	// Get duration
	duration, _ := audio.GetDuration(wavPath)

	// Transcribe
	ctx := context.Background()
	text, elapsed, err := a.transcriber.Transcribe(ctx, wavPath)
	if err != nil {
		log.Printf("[App] Transcription failed: %v", err)
		a.sound.Play(sound.Error)
		return
	}

	if text == "" {
		log.Println("[App] Empty transcription result (silence/hallucination)")
		a.sound.Play(sound.Success)
		return
	}

	// Save original clipboard
	var originalClip string
	if a.settings.RestoreClipboard {
		originalClip, _ = a.clipboard.GetText()
	}

	// Set text and paste
	if err := a.clipboard.SetText(text); err != nil {
		log.Printf("[App] Failed to set clipboard: %v", err)
		a.sound.Play(sound.Error)
		return
	}

	if err := a.clipboard.Paste(); err != nil {
		log.Printf("[App] Failed to paste: %v", err)
		a.sound.Play(sound.Error)
		return
	}

	a.sound.Play(sound.Success)

	// Restore clipboard after delay
	if a.settings.RestoreClipboard && originalClip != "" {
		go func() {
			time.Sleep(500 * time.Millisecond)
			a.clipboard.SetText(originalClip)
		}()
	}

	// Save to history
	wavData, err := os.ReadFile(wavPath)
	if err == nil {
		if _, err := history.Save(wavData, text, text, duration); err != nil {
			log.Printf("[App] Failed to save history: %v", err)
		}
	}

	log.Printf("[App] Done: %q (%.2fs)", text, elapsed)
}

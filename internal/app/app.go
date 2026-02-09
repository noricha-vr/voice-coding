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
	"github.com/noricha-vr/voicecode/internal/core/trace"
	"github.com/noricha-vr/voicecode/internal/core/transcriber"
	"github.com/noricha-vr/voicecode/internal/platform/clipboard"
	"github.com/noricha-vr/voicecode/internal/platform/hotkey"
	"github.com/noricha-vr/voicecode/internal/platform/overlay"
	"github.com/noricha-vr/voicecode/internal/platform/recorder"
	"github.com/noricha-vr/voicecode/internal/platform/sound"
	"github.com/noricha-vr/voicecode/internal/platform/tray"
)

const (
	audioSampleRateHz  = 16000
	trimSilenceEnvVar  = "VOICECODE_TRIM_SILENCE"
	minUsefulTrimDelta = 300 * time.Millisecond
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
	currentRun  *recordingRun

	processMu sync.Mutex // guards processRecording from concurrent execution
}

type recordingRun struct {
	tl                 *trace.Timeline
	recordingStartedAt time.Time
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
	// Set up settings callbacks before Run
	a.tray.SetSettingsCallbacks(tray.SettingsCallbacks{
		OnHotkeyChange:     a.onHotkeyChange,
		OnDurationChange:   a.onDurationChange,
		OnPushToTalkToggle: a.onPushToTalkToggle,
	})

	a.tray.Run(func() {
		// onReady
		log.Println("[App] Ready")
		a.tray.SetState(tray.Idle)

		// Sync tray menu with current settings
		a.tray.UpdateSettings(a.settings.Hotkey, a.settings.MaxRecordingDuration, a.settings.PushToTalk)

		a.registerHotkey()
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

func (a *App) registerHotkey() {
	if a.settings.PushToTalk {
		a.hotkey.Register(a.settings.Hotkey, a.onHotkeyPress, a.onHotkeyRelease)
	} else {
		a.hotkey.Register(a.settings.Hotkey, a.onHotkeyPress, nil)
	}
}

func (a *App) onHotkeyChange(key string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.hotkey.Unregister()
	a.settings.Hotkey = key
	a.registerHotkey()
	if err := a.settings.Save(); err != nil {
		log.Printf("[App] Failed to save settings: %v", err)
	}
	a.tray.UpdateSettings(a.settings.Hotkey, a.settings.MaxRecordingDuration, a.settings.PushToTalk)
	log.Printf("[App] Hotkey changed to: %s", key)
}

func (a *App) onDurationChange(seconds int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.settings.MaxRecordingDuration = seconds
	if err := a.settings.Save(); err != nil {
		log.Printf("[App] Failed to save settings: %v", err)
	}
	a.tray.UpdateSettings(a.settings.Hotkey, a.settings.MaxRecordingDuration, a.settings.PushToTalk)
	log.Printf("[App] Max recording duration changed to: %ds", seconds)
}

func (a *App) onPushToTalkToggle(enabled bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.settings.PushToTalk = enabled
	a.hotkey.Unregister()
	a.registerHotkey()
	if err := a.settings.Save(); err != nil {
		log.Printf("[App] Failed to save settings: %v", err)
	}
	a.tray.UpdateSettings(a.settings.Hotkey, a.settings.MaxRecordingDuration, a.settings.PushToTalk)
	log.Printf("[App] Push-to-talk: %v", enabled)
}

func (a *App) onHotkeyPress() {
	triggeredAt := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isRecording {
		a.stopAndProcess(triggeredAt)
	} else {
		a.startRecording(triggeredAt)
	}
}

func (a *App) onHotkeyRelease() {
	triggeredAt := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isRecording && a.settings.PushToTalk {
		a.stopAndProcess(triggeredAt)
	}
}

func (a *App) startRecording(triggeredAt time.Time) {
	tl := trace.NewWithStart("gui", triggeredAt)
	tl.Eventf("hotkey.start key=%s push_to_talk=%v max_recording_duration=%ds restore_clipboard=%v", a.settings.Hotkey, a.settings.PushToTalk, a.settings.MaxRecordingDuration, a.settings.RestoreClipboard)

	recStartDone := tl.Step("recorder.Start")
	if err := a.recorder.Start(); err != nil {
		recStartDone(err)
		log.Printf("[App] Failed to start recording: %v", err)
		a.sound.Play(sound.Error)
		tl.Finishf("aborted: recorder.Start failed")
		return
	}
	recStartDone(nil)

	a.isRecording = true
	a.currentRun = &recordingRun{tl: tl, recordingStartedAt: time.Now()}

	sndStartDone := tl.Step("sound.Play(Start)")
	sndStartDone(a.sound.Play(sound.Start))

	ovShowDone := tl.Step("overlay.Show")
	ovShowDone(a.overlay.Show("Recording..."))

	trayRecDone := tl.Step("tray.SetState(Recording)")
	trayRecDone(a.tray.SetState(tray.Recording))

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
				if a.currentRun != nil && a.currentRun.tl != nil {
					a.currentRun.tl.Eventf("recording.timeout max_recording_duration=%ds", a.settings.MaxRecordingDuration)
				}
				a.stopAndProcess(time.Now())
			}
		}
	}()
}

func (a *App) stopAndProcess(triggeredAt time.Time) {
	run := a.currentRun
	a.currentRun = nil

	var tl *trace.Timeline
	var talkDuration time.Duration
	if run != nil {
		tl = run.tl
		if !run.recordingStartedAt.IsZero() {
			talkDuration = triggeredAt.Sub(run.recordingStartedAt)
		}
	}

	if tl != nil {
		tl.Eventf("hotkey.stop talk_duration=%s", talkDuration.Truncate(time.Millisecond))
	}

	if a.cancelTimer != nil {
		a.cancelTimer()
		a.cancelTimer = nil
	}

	recStopDone := (*trace.Timeline)(nil)
	if tl != nil {
		recStopDone = tl
	}
	stopDone := recStopDone.Step("recorder.Stop")
	samples, err := a.recorder.Stop()
	stopDone(err)
	a.isRecording = false

	sndStopDone := recStopDone.Step("sound.Play(Stop)")
	sndStopDone(a.sound.Play(sound.Stop))

	ovHideDone := recStopDone.Step("overlay.Hide")
	ovHideDone(a.overlay.Hide())

	trayProcDone := recStopDone.Step("tray.SetState(Processing)")
	trayProcDone(a.tray.SetState(tray.Processing))

	if err != nil {
		log.Printf("[App] Failed to stop recording: %v", err)
		a.sound.Play(sound.Error)
		a.tray.SetState(tray.Idle)
		if tl != nil {
			tl.Finishf("aborted: recorder.Stop failed")
		}
		return
	}

	// Process in background to avoid blocking hotkey
	if tl != nil {
		tl.Eventf("processing.spawn samples=%d", len(samples))
	}
	go a.processRecording(samples, tl, talkDuration)
}

func (a *App) processRecording(samples []int16, tl *trace.Timeline, talkDuration time.Duration) {
	if tl != nil {
		tl.Eventf("processing.start samples=%d talk_duration=%s", len(samples), talkDuration.Truncate(time.Millisecond))
	}

	stepper := (*trace.Timeline)(nil)
	if tl != nil {
		stepper = tl
	}

	lockStart := time.Now()
	a.processMu.Lock()
	if tl != nil {
		tl.Eventf("processing.queue_wait=%s", time.Since(lockStart).Truncate(time.Millisecond))
	}
	defer a.processMu.Unlock()

	setProcDone := stepper.Step("tray.SetState(Processing)")
	setProcDone(a.tray.SetState(tray.Processing))
	defer func() {
		setIdleDone := stepper.Step("tray.SetState(Idle)")
		setIdleDone(a.tray.SetState(tray.Idle))
	}()

	// Save WAV to temp file
	tmpDir := os.TempDir()
	wavPath := filepath.Join(tmpDir, "voicecode_recording.wav")
	wavWriteDone := (*trace.Timeline)(nil)
	if tl != nil {
		wavWriteDone = tl
	}

	if envBoolDefaultTrue(trimSilenceEnvVar) {
		trimDone := stepper.Step("audio.TrimSilence")
		trimmed, info := audio.TrimSilence(samples, audioSampleRateHz)
		trimDone(nil)

		origDur := time.Duration(float64(info.OriginalSamples) / float64(audioSampleRateHz) * float64(time.Second)).Truncate(time.Millisecond)
		trimDur := time.Duration(float64(info.TrimmedSamples) / float64(audioSampleRateHz) * float64(time.Second)).Truncate(time.Millisecond)
		trimDelta := origDur - trimDur

		if tl != nil {
			tl.Eventf(
				"audio.trim_silence original=%s trimmed=%s lead=%d tail=%d window=%d pad=%d noise=%.1f threshold=%.1f all_silence=%v",
				origDur,
				trimDur,
				info.LeadingTrimmedSamples,
				info.TrailingTrimmedSamples,
				info.WindowSamples,
				info.PadSamples,
				info.NoiseFloor,
				info.Threshold,
				info.AllSilence,
			)
		}

		if info.AllSilence {
			log.Printf("[App] Silence detected. Skip transcription.")
			sndSuccessDone := stepper.Step("sound.Play(Success)")
			sndSuccessDone(a.sound.Play(sound.Success))
			if tl != nil {
				tl.Finishf("silence_skip original=%s", origDur)
			}
			return
		}

		if trimDelta >= minUsefulTrimDelta && len(trimmed) > 0 && len(trimmed) < len(samples) {
			samples = trimmed
			if tl != nil {
				tl.Eventf("audio.trim_silence applied delta=%s", trimDelta)
			}
		}
	}

	writeDone := wavWriteDone.Step("audio.WriteWAV")
	if err := audio.WriteWAV(wavPath, samples, audioSampleRateHz); err != nil {
		writeDone(err)
		log.Printf("[App] Failed to write WAV: %v", err)
		a.sound.Play(sound.Error)
		if tl != nil {
			tl.Finishf("aborted: audio.WriteWAV failed")
		}
		return
	}
	writeDone(nil)
	defer os.Remove(wavPath)

	// Get duration
	getDurDone := wavWriteDone.Step("audio.GetDuration")
	duration, durErr := audio.GetDuration(wavPath)
	getDurDone(durErr)

	// Transcribe
	if a.transcriber == nil {
		log.Printf("[App] Transcriber is nil")
		a.sound.Play(sound.Error)
		if tl != nil {
			tl.Finishf("aborted: transcriber is nil")
		}
		return
	}

	ctx := trace.WithTimeline(context.Background(), tl)
	txDone := wavWriteDone.Step("transcriber.Transcribe")
	text, elapsed, err := a.transcriber.Transcribe(ctx, wavPath)
	txDone(err)
	if err != nil {
		log.Printf("[App] Transcription failed: %v", err)
		a.sound.Play(sound.Error)
		if tl != nil {
			tl.Finishf("aborted: transcribe failed")
		}
		return
	}

	if text == "" {
		log.Println("[App] Empty transcription result (silence/hallucination)")
		a.sound.Play(sound.Success)
		if tl != nil {
			tl.Finishf("empty_result gemini_elapsed=%.2fs", elapsed)
		}
		return
	}

	// Save original clipboard
	var originalClip string
	if a.settings.RestoreClipboard {
		clipGetDone := wavWriteDone.Step("clipboard.GetText(original)")
		var clipErr error
		originalClip, clipErr = a.clipboard.GetText()
		clipGetDone(clipErr)
		if clipErr != nil {
			log.Printf("[App] Failed to get clipboard: %v", clipErr)
		}
	}

	// Set text and paste
	clipSetDone := wavWriteDone.Step("clipboard.SetText(result)")
	if err := a.clipboard.SetText(text); err != nil {
		clipSetDone(err)
		log.Printf("[App] Failed to set clipboard: %v", err)
		a.sound.Play(sound.Error)
		if tl != nil {
			tl.Finishf("aborted: clipboard.SetText failed")
		}
		return
	}
	clipSetDone(nil)

	pasteDone := wavWriteDone.Step("clipboard.Paste")
	if err := a.clipboard.Paste(); err != nil {
		pasteDone(err)
		log.Printf("[App] Failed to paste: %v", err)
		a.sound.Play(sound.Error)
		if tl != nil {
			tl.Finishf("aborted: clipboard.Paste failed")
		}
		return
	}
	pasteDone(nil)

	readyAt := time.Duration(0)
	if tl != nil {
		readyAt = tl.SinceStart()
		tl.Eventf("RESULT_READY")
	}

	sndSuccessDone := wavWriteDone.Step("sound.Play(Success)")
	sndSuccessDone(a.sound.Play(sound.Success))

	// Restore clipboard after delay
	if a.settings.RestoreClipboard && originalClip != "" {
		if tl != nil {
			tl.Eventf("async.clipboard.restore scheduled (500ms)")
		}
		go func() {
			time.Sleep(500 * time.Millisecond)
			if tl != nil {
				restoreDone := tl.Step("async.clipboard.SetText(restore)")
				restoreDone(a.clipboard.SetText(originalClip))
				return
			}
			a.clipboard.SetText(originalClip)
		}()
	}

	// Save to history
	readHistDone := wavWriteDone.Step("os.ReadFile(wav)")
	wavData, err := os.ReadFile(wavPath)
	readHistDone(err)
	if err == nil {
		saveHistDone := wavWriteDone.Step("history.Save")
		_, saveErr := history.Save(wavData, text, text, duration)
		saveHistDone(saveErr)
		if saveErr != nil {
			log.Printf("[App] Failed to save history: %v", saveErr)
		}
	}

	log.Printf("[App] Done: %q (%.2fs)", text, elapsed)
	if tl != nil {
		tl.Finishf("ok text_len=%d gemini_elapsed=%.2fs result_ready=%s", len(text), elapsed, readyAt.Truncate(time.Millisecond))
	}
}

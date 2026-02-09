package app

import (
	"testing"
	"time"

	"github.com/noricha-vr/voicecode/internal/core/settings"
	"github.com/noricha-vr/voicecode/internal/platform/sound"
	"github.com/noricha-vr/voicecode/internal/platform/tray"
)

// Mock implementations for testing

type mockRecorder struct {
	recording bool
	samples   []int16
}

func (m *mockRecorder) Start() error { m.recording = true; return nil }
func (m *mockRecorder) Stop() ([]int16, error) {
	m.recording = false
	return m.samples, nil
}
func (m *mockRecorder) IsRecording() bool { return m.recording }

type mockClipboard struct {
	text string
}

func (m *mockClipboard) GetText() (string, error) { return m.text, nil }
func (m *mockClipboard) SetText(t string) error   { m.text = t; return nil }
func (m *mockClipboard) Paste() error             { return nil }

type mockSound struct {
	lastPlayed sound.SoundType
}

func (m *mockSound) Play(s sound.SoundType) error { m.lastPlayed = s; return nil }

type mockOverlay struct {
	visible bool
	text    string
}

func (m *mockOverlay) Show(t string) error { m.visible = true; m.text = t; return nil }
func (m *mockOverlay) Hide() error         { m.visible = false; return nil }

type mockHotkey struct {
	onPress   func()
	onRelease func()
}

func (m *mockHotkey) Register(key string, onPress func(), onRelease func()) error {
	m.onPress = onPress
	m.onRelease = onRelease
	return nil
}
func (m *mockHotkey) Unregister() error { return nil }

type mockTray struct {
	state     tray.State
	ready     func()
	quit      func()
	cb        tray.SettingsCallbacks
	curHotkey string
	curDur    int
	curPTT    bool
}

func (m *mockTray) Run(onReady func(), onQuit func()) {
	m.ready = onReady
	m.quit = onQuit
	if onReady != nil {
		onReady()
	}
}
func (m *mockTray) SetState(state tray.State) error                { m.state = state; return nil }
func (m *mockTray) SetSettingsCallbacks(cb tray.SettingsCallbacks) { m.cb = cb }
func (m *mockTray) UpdateSettings(hotkey string, maxDuration int, pushToTalk bool) {
	m.curHotkey = hotkey
	m.curDur = maxDuration
	m.curPTT = pushToTalk
}

func TestNewApp(t *testing.T) {
	cfg := settings.Default()
	snd := &mockSound{}
	ov := &mockOverlay{}
	hk := &mockHotkey{}
	tr := &mockTray{}
	rec := &mockRecorder{}
	clip := &mockClipboard{}

	a := New(cfg, nil, rec, clip, snd, ov, hk, tr)
	if a == nil {
		t.Fatal("New returned nil")
	}
	if a.settings != cfg {
		t.Error("settings not set correctly")
	}
}

func TestSettingsCallbacksConnected(t *testing.T) {
	cfg := settings.Default()
	snd := &mockSound{}
	ov := &mockOverlay{}
	hk := &mockHotkey{}
	tr := &mockTray{}
	rec := &mockRecorder{}
	clip := &mockClipboard{}

	a := New(cfg, nil, rec, clip, snd, ov, hk, tr)
	a.Run()

	// Verify tray received settings callbacks
	if tr.cb.OnHotkeyChange == nil {
		t.Error("OnHotkeyChange callback not set")
	}
	if tr.cb.OnDurationChange == nil {
		t.Error("OnDurationChange callback not set")
	}
	if tr.cb.OnPushToTalkToggle == nil {
		t.Error("OnPushToTalkToggle callback not set")
	}

	// Verify initial settings synced to tray
	if tr.curHotkey != "f15" {
		t.Errorf("curHotkey = %q, want %q", tr.curHotkey, "f15")
	}
	if tr.curDur != 120 {
		t.Errorf("curDur = %d, want 120", tr.curDur)
	}
}

func TestOnHotkeyChange(t *testing.T) {
	cfg := settings.Default()
	hk := &mockHotkey{}
	tr := &mockTray{}
	a := New(cfg, nil, &mockRecorder{}, &mockClipboard{}, &mockSound{}, &mockOverlay{}, hk, tr)
	a.Run()

	// Simulate hotkey change from tray menu
	a.onHotkeyChange("f13")

	if a.settings.Hotkey != "f13" {
		t.Errorf("Hotkey = %q, want %q", a.settings.Hotkey, "f13")
	}
	if tr.curHotkey != "f13" {
		t.Errorf("tray curHotkey = %q, want %q", tr.curHotkey, "f13")
	}
}

func TestOnDurationChange(t *testing.T) {
	cfg := settings.Default()
	tr := &mockTray{}
	a := New(cfg, nil, &mockRecorder{}, &mockClipboard{}, &mockSound{}, &mockOverlay{}, &mockHotkey{}, tr)
	a.Run()

	a.onDurationChange(60)

	if a.settings.MaxRecordingDuration != 60 {
		t.Errorf("MaxRecordingDuration = %d, want 60", a.settings.MaxRecordingDuration)
	}
	if tr.curDur != 60 {
		t.Errorf("tray curDur = %d, want 60", tr.curDur)
	}
}

func TestOnPushToTalkToggle(t *testing.T) {
	cfg := settings.Default()
	tr := &mockTray{}
	a := New(cfg, nil, &mockRecorder{}, &mockClipboard{}, &mockSound{}, &mockOverlay{}, &mockHotkey{}, tr)
	a.Run()

	a.onPushToTalkToggle(true)

	if !a.settings.PushToTalk {
		t.Error("PushToTalk should be true")
	}
	if !tr.curPTT {
		t.Error("tray curPTT should be true")
	}
}

func TestStartStopRecording(t *testing.T) {
	cfg := settings.Default()
	rec := &mockRecorder{samples: make([]int16, 16000)}
	clip := &mockClipboard{}
	snd := &mockSound{}
	ov := &mockOverlay{}
	hk := &mockHotkey{}
	tr := &mockTray{}

	a := New(cfg, nil, rec, clip, snd, ov, hk, tr)

	// Start recording
	a.mu.Lock()
	a.startRecording(time.Now())
	a.mu.Unlock()

	if !a.isRecording {
		t.Error("expected isRecording to be true")
	}
	if !ov.visible {
		t.Error("expected overlay to be visible")
	}
	if snd.lastPlayed != sound.Start {
		t.Errorf("expected Start sound, got %v", snd.lastPlayed)
	}

	// Stop recording (will fail on transcribe since we have no transcriber, but state should change)
	a.mu.Lock()
	a.stopAndProcess(time.Now())
	a.mu.Unlock()

	if a.isRecording {
		t.Error("expected isRecording to be false")
	}
	if ov.visible {
		t.Error("expected overlay to be hidden")
	}
}

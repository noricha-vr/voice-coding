package app

import (
	"testing"

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
func (m *mockClipboard) Paste() error              { return nil }

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
	state tray.State
	ready func()
	quit  func()
}

func (m *mockTray) Run(onReady func(), onQuit func()) {
	m.ready = onReady
	m.quit = onQuit
	if onReady != nil {
		onReady()
	}
}
func (m *mockTray) SetState(state tray.State) error { m.state = state; return nil }

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
	a.startRecording()
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
	a.stopAndProcess()
	a.mu.Unlock()

	if a.isRecording {
		t.Error("expected isRecording to be false")
	}
	if ov.visible {
		t.Error("expected overlay to be hidden")
	}
}

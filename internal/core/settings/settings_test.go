package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func withTempSettingsPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".voicecode", "settings.json")
	settingsPathFunc = func() string { return path }
	t.Cleanup(func() { settingsPathFunc = defaultSettingsPath })
	return path
}

func TestDefault(t *testing.T) {
	s := Default()

	if s.Hotkey != DefaultHotkey {
		t.Errorf("Hotkey = %q, want %q", s.Hotkey, DefaultHotkey)
	}
	if s.RestoreClipboard != DefaultRestoreClipboard {
		t.Errorf("RestoreClipboard = %v, want %v", s.RestoreClipboard, DefaultRestoreClipboard)
	}
	if s.MaxRecordingDuration != DefaultMaxRecordingDuration {
		t.Errorf("MaxRecordingDuration = %d, want %d", s.MaxRecordingDuration, DefaultMaxRecordingDuration)
	}
	if s.PushToTalk != DefaultPushToTalk {
		t.Errorf("PushToTalk = %v, want %v", s.PushToTalk, DefaultPushToTalk)
	}
}

func TestLoadCreatesDefaultsWhenFileNotExist(t *testing.T) {
	path := withTempSettingsPath(t)

	s, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if s.Hotkey != DefaultHotkey {
		t.Errorf("Hotkey = %q, want %q", s.Hotkey, DefaultHotkey)
	}

	// File should have been created
	if _, err := os.Stat(path); err != nil {
		t.Errorf("settings file was not created: %v", err)
	}
}

func TestSaveAndReload(t *testing.T) {
	withTempSettingsPath(t)

	s := Default()
	s.Hotkey = "f13"
	s.MaxRecordingDuration = 60

	if err := s.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.Hotkey != "f13" {
		t.Errorf("Hotkey = %q, want %q", loaded.Hotkey, "f13")
	}
	if loaded.MaxRecordingDuration != 60 {
		t.Errorf("MaxRecordingDuration = %d, want 60", loaded.MaxRecordingDuration)
	}
}

func TestSaveCreatesValidJSON(t *testing.T) {
	path := withTempSettingsPath(t)

	s := Default()
	if err := s.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}
}

func TestValidateInRange(t *testing.T) {
	s := Default()
	if err := s.Validate(); err != nil {
		t.Errorf("Validate() error on defaults: %v", err)
	}

	s.MaxRecordingDuration = MinRecordingDuration
	if err := s.Validate(); err != nil {
		t.Errorf("Validate() error on min boundary: %v", err)
	}

	s.MaxRecordingDuration = MaxRecordingDuration
	if err := s.Validate(); err != nil {
		t.Errorf("Validate() error on max boundary: %v", err)
	}
}

func TestValidateOutOfRange(t *testing.T) {
	tests := []struct {
		name     string
		duration int
	}{
		{"below minimum", MinRecordingDuration - 1},
		{"above maximum", MaxRecordingDuration + 1},
		{"zero", 0},
		{"negative", -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Default()
			s.MaxRecordingDuration = tt.duration
			if err := s.Validate(); err == nil {
				t.Errorf("Validate() should fail for duration %d", tt.duration)
			}
		})
	}
}

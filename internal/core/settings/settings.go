package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	DefaultHotkey               = "f15"
	DefaultRestoreClipboard     = true
	DefaultMaxRecordingDuration = 120
	DefaultPushToTalk           = false
	MinRecordingDuration        = 10
	MaxRecordingDuration        = 300
)

// Settings holds user-configurable application settings.
type Settings struct {
	Hotkey               string `json:"hotkey"`
	RestoreClipboard     bool   `json:"restore_clipboard"`
	MaxRecordingDuration int    `json:"max_recording_duration"`
	PushToTalk           bool   `json:"push_to_talk"`
}

// Default returns a Settings with default values.
func Default() *Settings {
	return &Settings{
		Hotkey:               DefaultHotkey,
		RestoreClipboard:     DefaultRestoreClipboard,
		MaxRecordingDuration: DefaultMaxRecordingDuration,
		PushToTalk:           DefaultPushToTalk,
	}
}

// settingsPathFunc is overridable for testing.
var settingsPathFunc = defaultSettingsPath

func defaultSettingsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".voicecode", "settings.json")
}

// Load reads settings from ~/.voicecode/settings.json.
// If the file does not exist, it creates one with default values.
func Load() (*Settings, error) {
	path := settingsPathFunc()

	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("reading settings: %w", err)
		}
		s := Default()
		if err := s.Save(); err != nil {
			return nil, fmt.Errorf("creating default settings: %w", err)
		}
		return s, nil
	}

	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing settings: %w", err)
	}
	s.clampDefaults()
	return &s, nil
}

// Save writes the settings to ~/.voicecode/settings.json.
func (s *Settings) Save() error {
	path := settingsPathFunc()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating settings directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling settings: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing settings: %w", err)
	}
	return nil
}

// clampDefaults adjusts out-of-range values to the nearest valid boundary
// and logs warnings for each clamped field.
func (s *Settings) clampDefaults() {
	if s.MaxRecordingDuration < MinRecordingDuration {
		log.Printf("[Settings] max_recording_duration %d is below minimum %d, clamping", s.MaxRecordingDuration, MinRecordingDuration)
		s.MaxRecordingDuration = MinRecordingDuration
	}
	if s.MaxRecordingDuration > MaxRecordingDuration {
		log.Printf("[Settings] max_recording_duration %d is above maximum %d, clamping", s.MaxRecordingDuration, MaxRecordingDuration)
		s.MaxRecordingDuration = MaxRecordingDuration
	}
}

// Validate checks that settings values are within acceptable ranges.
func (s *Settings) Validate() error {
	if s.MaxRecordingDuration < MinRecordingDuration || s.MaxRecordingDuration > MaxRecordingDuration {
		return fmt.Errorf(
			"max_recording_duration must be between %d and %d, got %d",
			MinRecordingDuration, MaxRecordingDuration, s.MaxRecordingDuration,
		)
	}
	return nil
}

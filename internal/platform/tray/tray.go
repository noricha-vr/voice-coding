package tray

// State represents the current application state shown in the tray icon.
type State int

const (
	Idle State = iota
	Recording
	Processing
)

// SettingsCallbacks holds callbacks for settings changes from the tray menu.
type SettingsCallbacks struct {
	OnHotkeyChange     func(key string)
	OnDurationChange   func(seconds int)
	OnPushToTalkToggle func(enabled bool)
}

// Manager manages the system tray icon and menu.
type Manager interface {
	Run(onReady func(), onQuit func())
	SetState(state State) error
	SetSettingsCallbacks(cb SettingsCallbacks)
	UpdateSettings(hotkey string, maxDuration int, pushToTalk bool)
}

package tray

// State represents the current application state shown in the tray icon.
type State int

const (
	Idle State = iota
	Recording
	Processing
)

// Manager manages the system tray icon and menu.
type Manager interface {
	Run(onReady func(), onQuit func())
	SetState(state State) error
}

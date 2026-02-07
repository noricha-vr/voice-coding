package hotkey

// Manager registers and manages global hotkeys.
type Manager interface {
	Register(key string, onPress func(), onRelease func()) error
	Unregister() error
}

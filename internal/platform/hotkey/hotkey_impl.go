package hotkey

import (
	"fmt"
	"strings"

	xhotkey "golang.design/x/hotkey"
)

type hotkeyManager struct {
	hk *xhotkey.Hotkey
}

// NewManager creates a new global hotkey manager.
func NewManager() Manager {
	return &hotkeyManager{}
}

// parseKey parses "f15" or "ctrl+shift+r" into modifiers and key.
func parseKey(s string) ([]xhotkey.Modifier, xhotkey.Key, error) {
	parts := strings.Split(strings.ToLower(s), "+")
	var mods []xhotkey.Modifier
	keyStr := parts[len(parts)-1]
	for _, p := range parts[:len(parts)-1] {
		m, ok := modMap[p]
		if !ok {
			return nil, 0, fmt.Errorf("unknown modifier: %s", p)
		}
		mods = append(mods, m)
	}
	key, ok := keyMap[keyStr]
	if !ok {
		return nil, 0, fmt.Errorf("unknown key: %s", keyStr)
	}
	return mods, key, nil
}

func (m *hotkeyManager) Register(key string, onPress func(), onRelease func()) error {
	mods, k, err := parseKey(key)
	if err != nil {
		return err
	}
	m.hk = xhotkey.New(mods, k)
	if err := m.hk.Register(); err != nil {
		return fmt.Errorf("register hotkey: %w", err)
	}

	go func() {
		for range m.hk.Keydown() {
			if onPress != nil {
				onPress()
			}
		}
	}()
	go func() {
		for range m.hk.Keyup() {
			if onRelease != nil {
				onRelease()
			}
		}
	}()

	return nil
}

func (m *hotkeyManager) Unregister() error {
	if m.hk != nil {
		return m.hk.Unregister()
	}
	return nil
}

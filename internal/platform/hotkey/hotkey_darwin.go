//go:build darwin

package hotkey

import (
	"fmt"
	"strings"

	xhotkey "golang.design/x/hotkey"
)

var keyMap = map[string]xhotkey.Key{
	"f1": xhotkey.KeyF1, "f2": xhotkey.KeyF2, "f3": xhotkey.KeyF3, "f4": xhotkey.KeyF4,
	"f5": xhotkey.KeyF5, "f6": xhotkey.KeyF6, "f7": xhotkey.KeyF7, "f8": xhotkey.KeyF8,
	"f9": xhotkey.KeyF9, "f10": xhotkey.KeyF10, "f11": xhotkey.KeyF11, "f12": xhotkey.KeyF12,
	"f13": xhotkey.KeyF13, "f14": xhotkey.KeyF14, "f15": xhotkey.KeyF15,
	"f16": xhotkey.KeyF16, "f17": xhotkey.KeyF17, "f18": xhotkey.KeyF18,
	"f19": xhotkey.KeyF19, "f20": xhotkey.KeyF20,
	"a": xhotkey.KeyA, "b": xhotkey.KeyB, "c": xhotkey.KeyC, "d": xhotkey.KeyD,
	"e": xhotkey.KeyE, "f": xhotkey.KeyF, "g": xhotkey.KeyG, "h": xhotkey.KeyH,
	"i": xhotkey.KeyI, "j": xhotkey.KeyJ, "k": xhotkey.KeyK, "l": xhotkey.KeyL,
	"m": xhotkey.KeyM, "n": xhotkey.KeyN, "o": xhotkey.KeyO, "p": xhotkey.KeyP,
	"q": xhotkey.KeyQ, "r": xhotkey.KeyR, "s": xhotkey.KeyS, "t": xhotkey.KeyT,
	"u": xhotkey.KeyU, "v": xhotkey.KeyV, "w": xhotkey.KeyW, "x": xhotkey.KeyX,
	"y": xhotkey.KeyY, "z": xhotkey.KeyZ,
	"space": xhotkey.KeySpace,
}

var modMap = map[string]xhotkey.Modifier{
	"ctrl":  xhotkey.ModCtrl,
	"shift": xhotkey.ModShift,
	"alt":   xhotkey.ModOption,
	"cmd":   xhotkey.ModCmd,
}

type darwinManager struct {
	hk *xhotkey.Hotkey
}

// NewManager creates a new macOS global hotkey manager.
func NewManager() Manager {
	return &darwinManager{}
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

func (m *darwinManager) Register(key string, onPress func(), onRelease func()) error {
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

func (m *darwinManager) Unregister() error {
	if m.hk != nil {
		return m.hk.Unregister()
	}
	return nil
}

//go:build linux

package clipboard

import (
	"fmt"
	"os/exec"
	"time"

	xclip "golang.design/x/clipboard"
)

// NewClipboard creates a new Linux clipboard manager.
func NewClipboard() (Clipboard, error) {
	if err := xclip.Init(); err != nil {
		return nil, err
	}
	return &clipboardImpl{}, nil
}

func (c *clipboardImpl) Paste() error {
	time.Sleep(100 * time.Millisecond)

	// Try xdotool first (X11), then wtype (Wayland)
	if path, err := exec.LookPath("xdotool"); err == nil {
		cmd := exec.Command(path, "key", "--clearmodifiers", "ctrl+v")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("xdotool paste: %w", err)
		}
		return nil
	}

	if path, err := exec.LookPath("wtype"); err == nil {
		cmd := exec.Command(path, "-M", "ctrl", "-k", "v", "-m", "ctrl")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("wtype paste: %w", err)
		}
		return nil
	}

	return fmt.Errorf("no paste tool found: install xdotool (X11) or wtype (Wayland)")
}

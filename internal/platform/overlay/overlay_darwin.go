//go:build darwin

package overlay

import "log"

// darwinOverlay is a minimal overlay implementation.
// TODO: Implement NSWindow-based floating overlay.
type darwinOverlay struct{}

// NewOverlay creates a new macOS overlay (currently log-based stub).
func NewOverlay() Overlay {
	return &darwinOverlay{}
}

func (o *darwinOverlay) Show(text string) error {
	log.Printf("[Overlay] %s", text)
	return nil
}

func (o *darwinOverlay) Hide() error {
	return nil
}

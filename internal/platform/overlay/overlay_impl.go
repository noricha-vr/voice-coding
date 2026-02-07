package overlay

import "log"

// logOverlay is a minimal overlay implementation (log-based stub).
type logOverlay struct{}

// NewOverlay creates a new overlay (currently log-based stub).
func NewOverlay() Overlay {
	return &logOverlay{}
}

func (o *logOverlay) Show(text string) error {
	log.Printf("[Overlay] %s", text)
	return nil
}

func (o *logOverlay) Hide() error {
	return nil
}

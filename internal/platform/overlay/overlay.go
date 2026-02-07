package overlay

// Overlay displays a floating status indicator on screen.
type Overlay interface {
	Show(text string) error
	Hide() error
}

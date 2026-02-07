package clipboard

// Clipboard provides access to the system clipboard and paste simulation.
type Clipboard interface {
	GetText() (string, error)
	SetText(text string) error
	Paste() error
}

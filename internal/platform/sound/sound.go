package sound

// SoundType represents the type of system sound to play.
type SoundType int

const (
	Start SoundType = iota
	Stop
	Success
	Error
)

// Player plays system sounds for audio feedback.
type Player interface {
	Play(s SoundType) error
}

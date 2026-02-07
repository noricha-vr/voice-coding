//go:build darwin

package sound

import (
	"fmt"
	"os/exec"
)

var soundPaths = map[SoundType]string{
	Start:   "/System/Library/Sounds/Tink.aiff",
	Stop:    "/System/Library/Sounds/Pop.aiff",
	Success: "/System/Library/Sounds/Glass.aiff",
	Error:   "/System/Library/Sounds/Basso.aiff",
}

type darwinPlayer struct{}

// NewPlayer creates a new macOS sound player.
func NewPlayer() Player { return &darwinPlayer{} }

func (p *darwinPlayer) Play(s SoundType) error {
	path, ok := soundPaths[s]
	if !ok {
		return fmt.Errorf("unknown sound type: %d", s)
	}
	cmd := exec.Command("afplay", path)
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait() // reap child process to avoid zombie
	return nil
}

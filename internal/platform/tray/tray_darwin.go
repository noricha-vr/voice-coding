//go:build darwin

package tray

import (
	"fyne.io/systray"
	"github.com/noricha-vr/voicecode/assets"
)

type darwinManager struct {
	mQuit *systray.MenuItem
}

// NewManager creates a new macOS system tray manager.
func NewManager() Manager {
	return &darwinManager{}
}

func (m *darwinManager) Run(onReady func(), onQuit func()) {
	systray.Run(func() {
		systray.SetIcon(assets.IconIdle)
		systray.SetTitle("")
		systray.SetTooltip("VoiceCode")

		m.mQuit = systray.AddMenuItem("Quit", "Quit VoiceCode")

		go func() {
			<-m.mQuit.ClickedCh
			systray.Quit()
		}()

		if onReady != nil {
			onReady()
		}
	}, func() {
		if onQuit != nil {
			onQuit()
		}
	})
}

func (m *darwinManager) SetState(state State) error {
	switch state {
	case Idle:
		systray.SetIcon(assets.IconIdle)
	case Recording:
		systray.SetIcon(assets.IconRecording)
	case Processing:
		systray.SetIcon(assets.IconProcessing)
	}
	return nil
}

package tray

import (
	"fmt"

	"fyne.io/systray"
	"github.com/noricha-vr/voicecode/assets"
)

var hotkeyOptions = []string{"f13", "f14", "f15", "f16", "f17", "f18", "f19", "f20"}

type durationOption struct {
	label   string
	seconds int
}

var durationOptions = []durationOption{
	{"30s", 30},
	{"60s", 60},
	{"120s", 120},
	{"300s", 300},
}

type systrayManager struct {
	mQuit    *systray.MenuItem
	cb       SettingsCallbacks
	hkItems  []*systray.MenuItem
	durItems []*systray.MenuItem
	pttItem  *systray.MenuItem
}

// NewManager creates a new system tray manager.
func NewManager() Manager {
	return &systrayManager{}
}

func (m *systrayManager) SetSettingsCallbacks(cb SettingsCallbacks) {
	m.cb = cb
}

func (m *systrayManager) UpdateSettings(hotkey string, maxDuration int, pushToTalk bool) {
	for i, opt := range hotkeyOptions {
		if i < len(m.hkItems) {
			if opt == hotkey {
				m.hkItems[i].Check()
			} else {
				m.hkItems[i].Uncheck()
			}
		}
	}
	for i, opt := range durationOptions {
		if i < len(m.durItems) {
			if opt.seconds == maxDuration {
				m.durItems[i].Check()
			} else {
				m.durItems[i].Uncheck()
			}
		}
	}
	if m.pttItem != nil {
		if pushToTalk {
			m.pttItem.Check()
			m.pttItem.SetTitle("Push-to-Talk: On")
		} else {
			m.pttItem.Uncheck()
			m.pttItem.SetTitle("Push-to-Talk: Off")
		}
	}
}

func (m *systrayManager) Run(onReady func(), onQuit func()) {
	systray.Run(func() {
		systray.SetIcon(assets.IconIdle)
		systray.SetTitle("")
		systray.SetTooltip("VoiceCode")

		// Settings submenu
		mSettings := systray.AddMenuItem("Settings", "Application settings")

		// Hotkey submenu
		mHotkey := mSettings.AddSubMenuItem("Hotkey", "Change hotkey")
		m.hkItems = make([]*systray.MenuItem, len(hotkeyOptions))
		for i, key := range hotkeyOptions {
			m.hkItems[i] = mHotkey.AddSubMenuItemCheckbox(key, fmt.Sprintf("Use %s as hotkey", key), false)
		}

		// Duration submenu
		mDuration := mSettings.AddSubMenuItem("Max Duration", "Maximum recording duration")
		m.durItems = make([]*systray.MenuItem, len(durationOptions))
		for i, opt := range durationOptions {
			m.durItems[i] = mDuration.AddSubMenuItemCheckbox(opt.label, fmt.Sprintf("Record up to %s", opt.label), false)
		}

		// Push-to-Talk toggle
		m.pttItem = mSettings.AddSubMenuItemCheckbox("Push-to-Talk: Off", "Toggle push-to-talk mode", false)

		systray.AddSeparator()
		m.mQuit = systray.AddMenuItem("Quit", "Quit VoiceCode")

		// Listen for hotkey item clicks
		for i, key := range hotkeyOptions {
			go func(idx int, k string) {
				for range m.hkItems[idx].ClickedCh {
					if m.cb.OnHotkeyChange != nil {
						m.cb.OnHotkeyChange(k)
					}
				}
			}(i, key)
		}

		// Listen for duration item clicks
		for i, opt := range durationOptions {
			go func(idx int, secs int) {
				for range m.durItems[idx].ClickedCh {
					if m.cb.OnDurationChange != nil {
						m.cb.OnDurationChange(secs)
					}
				}
			}(i, opt.seconds)
		}

		// Listen for push-to-talk and quit
		go func() {
			for {
				select {
				case <-m.mQuit.ClickedCh:
					systray.Quit()
					return
				case <-m.pttItem.ClickedCh:
					enabled := !m.pttItem.Checked()
					if m.cb.OnPushToTalkToggle != nil {
						m.cb.OnPushToTalkToggle(enabled)
					}
				}
			}
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

func (m *systrayManager) SetState(state State) error {
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

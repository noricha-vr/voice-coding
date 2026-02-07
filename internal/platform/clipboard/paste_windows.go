//go:build windows

package clipboard

import (
	"fmt"
	"log"
	"time"
	"unsafe"

	xclip "golang.design/x/clipboard"

	"golang.org/x/sys/windows"
)

var (
	user32              = windows.NewLazySystemDLL("user32.dll")
	procSendInput       = user32.NewProc("SendInput")
	procMapVirtualKeyW  = user32.NewProc("MapVirtualKeyW")
)

const (
	inputKeyboard   = 1
	keventfKeyup    = 0x0002
	keventfScancode = 0x0008
	vkControl       = 0x11
	vkV             = 0x56
)

type keyboardInput struct {
	typ uint32
	ki  struct {
		wVk         uint16
		wScan       uint16
		dwFlags     uint32
		time        uint32
		dwExtraInfo uintptr
	}
	padding [8]byte
}

func sendKey(vk uint16, flags uint32) {
	scan, _, _ := procMapVirtualKeyW.Call(uintptr(vk), 0)
	input := keyboardInput{typ: inputKeyboard}
	input.ki.wVk = vk
	input.ki.wScan = uint16(scan)
	input.ki.dwFlags = flags | keventfScancode
	ret, _, _ := procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), unsafe.Sizeof(input))
	if ret == 0 {
		log.Printf("[Clipboard] SendInput failed for vk=0x%X flags=0x%X", vk, flags)
	}
}

// NewClipboard creates a new Windows clipboard manager.
func NewClipboard() (Clipboard, error) {
	if err := xclip.Init(); err != nil {
		return nil, err
	}
	if err := procSendInput.Find(); err != nil {
		return nil, fmt.Errorf("SendInput not available: %w", err)
	}
	return &clipboardImpl{}, nil
}

func (c *clipboardImpl) Paste() error {
	time.Sleep(100 * time.Millisecond)
	sendKey(vkControl, 0)
	sendKey(vkV, 0)
	sendKey(vkV, keventfKeyup)
	sendKey(vkControl, keventfKeyup)
	return nil
}

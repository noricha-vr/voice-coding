//go:build darwin

package clipboard

/*
#cgo LDFLAGS: -framework ApplicationServices -framework Carbon
#include <ApplicationServices/ApplicationServices.h>
#include <Carbon/Carbon.h>

void simulatePaste() {
    CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
    CGEventRef keyDown = CGEventCreateKeyboardEvent(source, (CGKeyCode)9, true);   // 9 = 'v'
    CGEventRef keyUp = CGEventCreateKeyboardEvent(source, (CGKeyCode)9, false);
    CGEventSetFlags(keyDown, kCGEventFlagMaskCommand);
    CGEventSetFlags(keyUp, kCGEventFlagMaskCommand);
    CGEventPost(kCGAnnotatedSessionEventTap, keyDown);
    CGEventPost(kCGAnnotatedSessionEventTap, keyUp);
    CFRelease(keyDown);
    CFRelease(keyUp);
    CFRelease(source);
}
*/
import "C"

import (
	"time"

	xclip "golang.design/x/clipboard"
)

// NewClipboard creates a new macOS clipboard manager.
func NewClipboard() (Clipboard, error) {
	if err := xclip.Init(); err != nil {
		return nil, err
	}
	return &clipboardImpl{}, nil
}

func (c *clipboardImpl) Paste() error {
	time.Sleep(100 * time.Millisecond)
	C.simulatePaste()
	return nil
}

//go:build windows

package hotkey

import xhotkey "golang.design/x/hotkey"

var keyMap = map[string]xhotkey.Key{
	"f1": xhotkey.KeyF1, "f2": xhotkey.KeyF2, "f3": xhotkey.KeyF3, "f4": xhotkey.KeyF4,
	"f5": xhotkey.KeyF5, "f6": xhotkey.KeyF6, "f7": xhotkey.KeyF7, "f8": xhotkey.KeyF8,
	"f9": xhotkey.KeyF9, "f10": xhotkey.KeyF10, "f11": xhotkey.KeyF11, "f12": xhotkey.KeyF12,
	"f13": xhotkey.KeyF13, "f14": xhotkey.KeyF14, "f15": xhotkey.KeyF15,
	"f16": xhotkey.KeyF16, "f17": xhotkey.KeyF17, "f18": xhotkey.KeyF18,
	"f19": xhotkey.KeyF19, "f20": xhotkey.KeyF20,
	"a": xhotkey.KeyA, "b": xhotkey.KeyB, "c": xhotkey.KeyC, "d": xhotkey.KeyD,
	"e": xhotkey.KeyE, "f": xhotkey.KeyF, "g": xhotkey.KeyG, "h": xhotkey.KeyH,
	"i": xhotkey.KeyI, "j": xhotkey.KeyJ, "k": xhotkey.KeyK, "l": xhotkey.KeyL,
	"m": xhotkey.KeyM, "n": xhotkey.KeyN, "o": xhotkey.KeyO, "p": xhotkey.KeyP,
	"q": xhotkey.KeyQ, "r": xhotkey.KeyR, "s": xhotkey.KeyS, "t": xhotkey.KeyT,
	"u": xhotkey.KeyU, "v": xhotkey.KeyV, "w": xhotkey.KeyW, "x": xhotkey.KeyX,
	"y": xhotkey.KeyY, "z": xhotkey.KeyZ,
	"space": xhotkey.KeySpace,
}

var modMap = map[string]xhotkey.Modifier{
	"ctrl":  xhotkey.ModCtrl,
	"shift": xhotkey.ModShift,
	"alt":   xhotkey.ModAlt,
	"cmd":   xhotkey.ModWin,
}

package browser

import (
	"github.com/inkyblackness/shocked-client/env/keys"
)

// Mapping of key names to the values.
// Refer to https://developer.mozilla.org/en-US/docs/Web/API/KeyboardEvent/key/Key_Values
var keyMap = map[string]keys.Key{

	"Enter":     keys.KeyEnter,
	"Esc":       keys.KeyEscape,
	"Escape":    keys.KeyEscape,
	"Backspace": keys.KeyBackspace,
	"Tab":       keys.KeyTab,

	"ArrowDown":  keys.KeyDown,
	"ArrowLeft":  keys.KeyLeft,
	"ArrowRight": keys.KeyRight,
	"ArrowUp":    keys.KeyUp,
	"Down":       keys.KeyDown,
	"Left":       keys.KeyLeft,
	"Right":      keys.KeyRight,
	"Up":         keys.KeyUp,

	"Del":      keys.KeyDelete,
	"Delete":   keys.KeyDelete,
	"End":      keys.KeyEnd,
	"Home":     keys.KeyHome,
	"Insert":   keys.KeyInsert,
	"PageDown": keys.KeyPageDown,
	"PageUp":   keys.KeyPageUp,

	"Alt":        keys.KeyAlt,
	"AltGraph":   keys.KeyAlt,
	"ModeChange": keys.KeyAlt,
	"Control":    keys.KeyControl,
	"Shift":      keys.KeyShift,
	"Super":      keys.KeySuper,

	"Pause":       keys.KeyPause,
	"PrintScreen": keys.KeyPrintScreen,

	"CapsLock":   keys.KeyCapsLock,
	"Scroll":     keys.KeyScrollLock,
	"ScrollLock": keys.KeyScrollLock,

	"F1":  keys.KeyF1,
	"F10": keys.KeyF10,
	"F11": keys.KeyF11,
	"F12": keys.KeyF12,
	"F13": keys.KeyF13,
	"F14": keys.KeyF14,
	"F15": keys.KeyF15,
	"F16": keys.KeyF16,
	"F17": keys.KeyF17,
	"F18": keys.KeyF18,
	"F19": keys.KeyF19,
	"F2":  keys.KeyF2,
	"F20": keys.KeyF20,
	"F21": keys.KeyF21,
	"F22": keys.KeyF22,
	"F23": keys.KeyF23,
	"F24": keys.KeyF24,
	"F25": keys.KeyF25,
	"F3":  keys.KeyF3,
	"F4":  keys.KeyF4,
	"F5":  keys.KeyF5,
	"F6":  keys.KeyF6,
	"F7":  keys.KeyF7,
	"F8":  keys.KeyF8,
	"F9":  keys.KeyF9,

	"Copy":  keys.KeyCopy,
	"Cut":   keys.KeyCut,
	"Paste": keys.KeyPaste,

	"Undo": keys.KeyUndo,
	"Redo": keys.KeyRedo}

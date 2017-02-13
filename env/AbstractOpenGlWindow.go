package env

import (
	"github.com/inkyblackness/shocked-client/env/keys"
)

type keyDeferrer struct {
	window *AbstractOpenGlWindow
}

func (def *keyDeferrer) KeyDown(key keys.Key, modifier keys.Modifier) {
	def.window.CallKeyDown(key, modifier)
}

func (def *keyDeferrer) KeyUp(key keys.Key, modifier keys.Modifier) {
	def.window.CallKeyUp(key, modifier)
}

// AbstractOpenGlWindow implements the common, basic functionality of OpenGlWindow
type AbstractOpenGlWindow struct {
	keyBuffer *keys.StickyKeyBuffer

	CallRender            RenderCallback
	CallResize            ResizeCallback
	CallOnMouseMove       MouseMoveCallback
	CallOnMouseButtonUp   MouseButtonCallback
	CallOnMouseButtonDown MouseButtonCallback
	CallOnMouseScroll     MouseScrollCallback
	CallKeyUp             KeyCallback
	CallKeyDown           KeyCallback
	CallCharCallback      CharCallback
}

// InitAbstractOpenGlWindow returns an initialized instance.
func InitAbstractOpenGlWindow() AbstractOpenGlWindow {
	return AbstractOpenGlWindow{
		CallRender:            func() {},
		CallResize:            func(int, int) {},
		CallOnMouseMove:       func(float32, float32) {},
		CallOnMouseButtonUp:   func(uint32, keys.Modifier) {},
		CallOnMouseButtonDown: func(uint32, keys.Modifier) {},
		CallKeyDown:           func(keys.Key, keys.Modifier) {},
		CallKeyUp:             func(keys.Key, keys.Modifier) {},
		CallCharCallback:      func(rune) {}}
}

// StickyKeyListener returns an instance of a listener acting as an adapter
// for the key-down/-up callbacks.
func (window *AbstractOpenGlWindow) StickyKeyListener() keys.StickyKeyListener {
	return &keyDeferrer{window}
}

// OnRender implements the OpenGlWindow interface.
func (window *AbstractOpenGlWindow) OnRender(callback RenderCallback) {
	window.CallRender = callback
}

// OnResize implements the OpenGlWindow interface.
func (window *AbstractOpenGlWindow) OnResize(callback ResizeCallback) {
	window.CallResize = callback
}

// OnMouseMove implements the OpenGlWindow interface.
func (window *AbstractOpenGlWindow) OnMouseMove(callback MouseMoveCallback) {
	window.CallOnMouseMove = callback
}

// OnMouseButtonDown implements the OpenGlWindow interface.
func (window *AbstractOpenGlWindow) OnMouseButtonDown(callback MouseButtonCallback) {
	window.CallOnMouseButtonDown = callback
}

// OnMouseButtonUp implements the OpenGlWindow interface.
func (window *AbstractOpenGlWindow) OnMouseButtonUp(callback MouseButtonCallback) {
	window.CallOnMouseButtonUp = callback
}

// OnMouseScroll implements the OpenGlWindow interface.
func (window *AbstractOpenGlWindow) OnMouseScroll(callback MouseScrollCallback) {
	window.CallOnMouseScroll = callback
}

// OnKeyDown implements the OpenGlWindow interface
func (window *AbstractOpenGlWindow) OnKeyDown(callback KeyCallback) {
	window.CallKeyDown = callback
}

// OnKeyUp implements the OpenGlWindow interface
func (window *AbstractOpenGlWindow) OnKeyUp(callback KeyCallback) {
	window.CallKeyUp = callback
}

// OnCharCallback implements the OpenGlWindow interface
func (window *AbstractOpenGlWindow) OnCharCallback(callback CharCallback) {
	window.CallCharCallback = callback
}

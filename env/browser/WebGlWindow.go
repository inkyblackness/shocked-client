package browser

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/webgl"

	"github.com/inkyblackness/shocked-client/opengl"
)

// WebGlWindow represents a WebGL surface.
type WebGlWindow struct {
	canvas    *js.Object
	glWrapper *WebGl

	render func()
}

// NewWebGlWindow tries to initialize the WebGL environment and returns a
// new window instance.
func NewWebGlWindow(canvas *js.Object) (window *WebGlWindow, err error) {
	attrs := webgl.DefaultAttributes()
	attrs.Alpha = false

	glContext, err := webgl.NewContext(canvas, attrs)
	if err == nil {
		window = &WebGlWindow{
			canvas:    canvas,
			glWrapper: NewWebGl(glContext),
			render:    func() {}}

		browserWindow := js.Global.Get("window")
		type indirecterType struct {
			render func()
		}
		var indirecter indirecterType

		indirecter.render = func() {
			browserWindow.Call("requestAnimationFrame", indirecter.render)
			window.render()
		}
		indirecter.render()
	}

	return
}

// OpenGl implements the env.OpenGlWindow interface.
func (window *WebGlWindow) OpenGl() opengl.OpenGl {
	return window.glWrapper
}

// OnRender implements the env.OpenGlWindow interface.
func (window *WebGlWindow) OnRender(callback func()) {
	window.render = callback
}

// Size implements the env.OpenGlWindow interface.
func (window *WebGlWindow) Size() (width int, height int) {
	canvasWidth := window.canvas.Get("width").Int()
	canvasHeight := window.canvas.Get("height").Int()

	width = window.canvas.Get("clientWidth").Int()
	height = window.canvas.Get("clientHeight").Int()

	if canvasWidth != width {
		fmt.Printf("Setting canvas width %d to reported width %d\n", canvasWidth, width)
		window.canvas.Set("width", width)
	}
	if canvasHeight != height {
		fmt.Printf("Setting canvas height %d to reported height %d\n", canvasHeight, height)
		window.canvas.Set("height", height)
	}

	return
}

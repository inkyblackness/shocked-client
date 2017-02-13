package browser

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/webgl"

	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/env/keys"
	"github.com/inkyblackness/shocked-client/opengl"
)

var buttonsByIndex = map[int]uint32{
	0: env.MousePrimary,
	2: env.MouseSecondary}

// WebGlWindow represents a WebGL surface.
type WebGlWindow struct {
	env.AbstractOpenGlWindow

	keyBuffer *keys.StickyKeyBuffer

	canvas    *js.Object
	glWrapper *WebGl
}

// NewWebGlWindow tries to initialize the WebGL environment and returns a
// new window instance.
func NewWebGlWindow(canvas *js.Object) (window *WebGlWindow, err error) {
	attrs := webgl.DefaultAttributes()
	attrs.Alpha = false

	glContext, err := webgl.NewContext(canvas, attrs)
	if err == nil {
		window = &WebGlWindow{
			AbstractOpenGlWindow: env.InitAbstractOpenGlWindow(),

			canvas:    canvas,
			glWrapper: NewWebGl(glContext)}

		window.keyBuffer = keys.NewStickyKeyBuffer(window.StickyKeyListener())

		window.registerMouseListener()
		window.startRenderLoop()
	}

	return
}

func (window *WebGlWindow) registerMouseListener() {
	browserWindow := js.Global.Get("window")
	notifiedMouseButtons := uint32(0)

	getEventPosition := func(event *js.Object) (float64, float64, bool) {
		rect := window.canvas.Call("getBoundingClientRect")
		clientX := event.Get("clientX").Float()
		clientY := event.Get("clientY").Float()
		x := clientX - rect.Get("left").Float()
		y := clientY - rect.Get("top").Float()
		inRect := false

		if (x >= 0) && (clientX <= rect.Get("right").Float()) &&
			(y >= 0) && (clientY <= rect.Get("bottom").Float()) {
			inRect = true
		}

		return x, y, inRect
	}
	getEventModifier := func(event *js.Object) keys.Modifier {
		modifier := keys.ModNone

		if event.Get("ctrlKey").Bool() {
			modifier = modifier.With(keys.ModControl)
		}
		if event.Get("shiftKey").Bool() {
			modifier = modifier.With(keys.ModShift)
		}
		if event.Get("altKey").Bool() {
			modifier = modifier.With(keys.ModAlt)
		}
		if event.Get("metaKey").Bool() {
			modifier = modifier.With(keys.ModSuper)
		}

		return modifier
	}

	window.canvas.Call("addEventListener", "contextmenu", func(event *js.Object) bool {
		event.Call("preventDefault")
		return false
	})
	browserWindow.Call("addEventListener", "mousemove", func(event *js.Object) {
		x, y, inRect := getEventPosition(event)

		if inRect || (notifiedMouseButtons != 0) {
			window.CallOnMouseMove(float32(x), float32(y))
		}
	})
	browserWindow.Call("addEventListener", "mousedown", func(event *js.Object) {
		button, knownButton := buttonsByIndex[event.Get("button").Int()]

		if knownButton {
			_, _, inRect := getEventPosition(event)

			if (inRect || (notifiedMouseButtons != 0)) && ((notifiedMouseButtons & button) != button) {
				notifiedMouseButtons |= button
				modifier := getEventModifier(event)
				window.CallOnMouseButtonDown(button, modifier)
			}
		}
	})
	browserWindow.Call("addEventListener", "mouseup", func(event *js.Object) {
		button, knownButton := buttonsByIndex[event.Get("button").Int()]

		if knownButton {
			if (notifiedMouseButtons & button) == button {
				notifiedMouseButtons &= ^button
				modifier := getEventModifier(event)
				window.CallOnMouseButtonUp(button, modifier)
			}
		}
	})
	window.canvas.Call("addEventListener", "wheel", func(event *js.Object) {
		_, _, inRect := getEventPosition(event)

		if inRect || (notifiedMouseButtons != 0) {
			dx := event.Get("deltaX").Float()
			dy := event.Get("deltaY").Float()

			event.Call("preventDefault")
			window.CallOnMouseScroll(float32(dx), float32(dy))
		}
	})
	js.Global.Call("addEventListener", "keydown", func(event *js.Object) {
		key, knownKey := keyMap[event.Get("key").String()]

		if knownKey {
			event.Call("preventDefault")
			modifier := getEventModifier(event)
			window.keyBuffer.KeyDown(key, modifier)
		}
	})
	js.Global.Call("addEventListener", "keyup", func(event *js.Object) {
		key, knownKey := keyMap[event.Get("key").String()]

		if knownKey {
			event.Call("preventDefault")
			modifier := getEventModifier(event)
			window.keyBuffer.KeyUp(key, modifier)
		}
	})
	js.Global.Call("addEventListener", "blur", func(event *js.Object) {
		window.keyBuffer.ReleaseAll()
	})
	js.Global.Call("addEventListener", "keypress", func(event *js.Object) {
		charCode := event.Get("charCode").Int()
		charMember := event.Get("char")
		chars := ""

		if charMember != js.Undefined {
			chars = charMember.String()
		}
		if len(chars) > 0 {
			for _, char := range chars {
				window.CallCharCallback(char)
			}
		} else if charCode != 0 {
			window.CallCharCallback(rune(charCode))
		}
	})
}

func (window *WebGlWindow) startRenderLoop() {
	type indirecterType struct {
		render func()
	}
	var indirecter indirecterType
	browserWindow := js.Global.Get("window")
	lastWidth, lastHeight := window.Size()

	indirecter.render = func() {
		browserWindow.Call("requestAnimationFrame", indirecter.render)
		curWidth, curHeight := window.Size()

		if (curWidth != lastWidth) || (curHeight != lastHeight) {
			lastWidth, lastHeight = curWidth, curHeight
			window.CallResize(curWidth, curHeight)
		}
		window.CallRender()
	}
	indirecter.render()
}

// OpenGl implements the env.OpenGlWindow interface.
func (window *WebGlWindow) OpenGl() opengl.OpenGl {
	return window.glWrapper
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

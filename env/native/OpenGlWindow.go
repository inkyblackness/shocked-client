package native

import (
	"time"

	"github.com/go-gl/glfw/v3.1/glfw"

	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/opengl"
)

var framesPerSecond = float64(30.0)
var frameTime = time.Duration(int64(float64(time.Second) / float64(framesPerSecond)))

var buttonsByIndex = map[glfw.MouseButton]uint32{
	glfw.MouseButton1: env.MousePrimary,
	glfw.MouseButton2: env.MouseSecondary}

// OpenGlWindow represents a native OpenGL surface.
type OpenGlWindow struct {
	env.AbstractOpenGlWindow

	glfwWindow *glfw.Window
	glWrapper  *OpenGl

	nextRenderTick time.Time
}

// NewOpenGlWindow tries to initialize the OpenGL environment and returns a
// new window instance.
func NewOpenGlWindow() (window *OpenGlWindow, err error) {
	if err = glfw.Init(); err == nil {
		glfw.WindowHint(glfw.Resizable, 1)
		glfw.WindowHint(glfw.Decorated, 1)
		glfw.WindowHint(glfw.ClientAPI, glfw.OpenGLAPI)
		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 2)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
		var glfwWindow *glfw.Window
		glfwWindow, err = glfw.CreateWindow(320, 200, "shocked", nil, nil)
		if err == nil {
			glfwWindow.MakeContextCurrent()

			window = &OpenGlWindow{
				AbstractOpenGlWindow: env.InitAbstractOpenGlWindow(),
				glfwWindow:           glfwWindow,
				glWrapper:            NewOpenGl(),
				nextRenderTick:       time.Now()}

			glfwWindow.SetCursorPosCallback(window.onCursorPos)
			glfwWindow.SetMouseButtonCallback(window.onMouseButton)
			glfwWindow.SetScrollCallback(window.onMouseScroll)
			glfwWindow.SetFramebufferSizeCallback(window.onFramebufferResize)
			glfwWindow.SetCharCallback(window.onChar)
		}
	}
	return
}

// ShouldClose returns true if the user requested the window to close.
func (window *OpenGlWindow) ShouldClose() bool {
	return window.glfwWindow.ShouldClose()
}

// Close closes the window and releases its resources.
func (window *OpenGlWindow) Close() {
	window.glfwWindow.Destroy()
	glfw.Terminate()
}

// Update must be called from within the main thread as often as possible.
func (window *OpenGlWindow) Update() {
	glfw.PollEvents()

	now := time.Now()
	delta := now.Sub(window.nextRenderTick)
	if delta.Nanoseconds() < 0 {
		// detected a change of wallclock time into the past; realign
		delta = frameTime
		window.nextRenderTick = now
	}

	if delta.Nanoseconds() >= frameTime.Nanoseconds() {
		window.glfwWindow.MakeContextCurrent()
		window.CallRender()
		window.glfwWindow.SwapBuffers()
		framesCovered := delta.Nanoseconds() / frameTime.Nanoseconds()
		window.nextRenderTick = window.nextRenderTick.Add(time.Duration(framesCovered * frameTime.Nanoseconds()))
	}
}

// OpenGl implements the env.OpenGlWindow interface.
func (window *OpenGlWindow) OpenGl() opengl.OpenGl {
	return window.glWrapper
}

// Size implements the env.OpenGlWindow interface.
func (window *OpenGlWindow) Size() (width int, height int) {
	return window.glfwWindow.GetFramebufferSize()
}

func (window *OpenGlWindow) onFramebufferResize(rawWindow *glfw.Window, width int, height int) {
	window.CallResize(width, height)
}

func (window *OpenGlWindow) onCursorPos(rawWindow *glfw.Window, x float64, y float64) {
	window.CallOnMouseMove(float32(x), float32(y))
}

func (window *OpenGlWindow) onMouseButton(rawWindow *glfw.Window, rawButton glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	button, knownButton := buttonsByIndex[rawButton]

	if knownButton {
		modifierMask := uint32(0)

		if (mods & glfw.ModControl) != 0 {
			modifierMask |= env.ModControl
		}
		if (mods & glfw.ModShift) != 0 {
			modifierMask |= env.ModShift
		}
		if action == glfw.Press {
			window.CallOnMouseButtonDown(button, modifierMask)
		} else if action == glfw.Release {
			window.CallOnMouseButtonUp(button, modifierMask)
		}
	}
}

func (window *OpenGlWindow) onMouseScroll(rawWindow *glfw.Window, dx float64, dy float64) {
	window.CallOnMouseScroll(float32(dx), float32(dy)*-1.0)
}

func (window *OpenGlWindow) onChar(rawWindow *glfw.Window, char rune) {
	window.CallCharCallback(char)
}

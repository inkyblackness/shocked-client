package native

import (
	"github.com/go-gl/glfw/v3.1/glfw"

	"github.com/inkyblackness/shocked-client/opengl"
)

// OpenGlWindow represents a native OpenGL surface.
type OpenGlWindow struct {
	glfwWindow *glfw.Window
	glWrapper  *OpenGl

	render func()
}

// NewOpenGlWindow tries to initialize the OpenGL environment and returns a
// new window instance.
func NewOpenGlWindow() (window *OpenGlWindow, err error) {
	if err = glfw.Init(); err == nil {
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
				glfwWindow: glfwWindow,
				glWrapper:  NewOpenGl(),
				render:     func() {}}
		}
	}
	return
}

// Close closes the window and releases its resources.
func (window *OpenGlWindow) Close() {
	window.glfwWindow.Destroy()
	glfw.Terminate()
}

// Update must be called from within the main thread as often as possible.
func (window *OpenGlWindow) Update() {
	glfw.PollEvents()

	window.glfwWindow.MakeContextCurrent()
	window.render()
	window.glfwWindow.SwapBuffers()
}

// OpenGl implements the env.OpenGlWindow interface.
func (window *OpenGlWindow) OpenGl() opengl.OpenGl {
	return window.glWrapper
}

// OnRender implements the env.OpenGlWindow interface.
func (window *OpenGlWindow) OnRender(callback func()) {
	window.render = callback
}

// Size implements the env.OpenGlWindow interface.
func (window *OpenGlWindow) Size() (width int, height int) {
	return window.glfwWindow.GetSize()
}

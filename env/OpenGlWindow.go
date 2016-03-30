package env

import (
	"github.com/inkyblackness/shocked-client/opengl"
)

// OpenGlWindow represents an OpenGL render surface
type OpenGlWindow interface {
	// OpenGl returns the OpenGL API wrapper for this window.
	OpenGl() opengl.OpenGl
	// OnRender registers a callback function which shall be called to update the scene.
	OnRender(callback func())

	// Size returns the dimensions of the window in pixel.
	Size() (width float32, height float32)
	// OnResize registers a callback function which shall be called when the window is resized.
	OnResize(callback func())
}

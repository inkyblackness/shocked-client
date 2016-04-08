package env

import (
	"github.com/inkyblackness/shocked-client/opengl"
)

// RenderCallback is the function to receive render events.
type RenderCallback func()

// MouseMoveCallback is the function to receive the current mouse coordinate while moving.
type MouseMoveCallback func(x float32, y float32)

// MouseButtonCallback is the function to receive button up/down events.
type MouseButtonCallback func(buttonMask uint32)

// OpenGlWindow represents an OpenGL render surface
type OpenGlWindow interface {
	// OpenGl returns the OpenGL API wrapper for this window.
	OpenGl() opengl.OpenGl
	// OnRender registers a callback function which shall be called to update the scene.
	OnRender(callback RenderCallback)

	// Size returns the dimensions of the window display area in pixel.
	Size() (width int, height int)

	// OnMouseMove registers a callback function for mouse move events.
	OnMouseMove(callback MouseMoveCallback)
	// OnMouseButtonDown registers a callback function for mouse button down events.
	OnMouseButtonDown(callback MouseButtonCallback)
	// OnMouseButtonUp registers a callback function for mouse button up events.
	OnMouseButtonUp(callback MouseButtonCallback)
}

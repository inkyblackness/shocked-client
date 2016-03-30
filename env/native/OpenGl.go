package native

import (
	"github.com/go-gl/gl/v3.2-core/gl"
)

// OpenGl wraps the native GL API into a common interface
type OpenGl struct {
}

// NewOpenGl initializes the Gl bindings and returns an OpenGl instance.
func NewOpenGl() *OpenGl {
	opengl := &OpenGl{}

	if err := gl.Init(); err != nil {
		panic(err)
	}

	return opengl
}

// GetError implements the opengl.OpenGl interface.
func (openGl *OpenGl) GetError() uint32 {
	return gl.GetError()
}

// Clear implements the opengl.OpenGl interface.
func (openGl *OpenGl) Clear(mask uint32) {
	gl.Clear(mask)
}

// ClearColor implements the opengl.OpenGl interface.
func (openGl *OpenGl) ClearColor(red, green, blue, alpha float32) {
	gl.ClearColor(red, green, blue, alpha)
}

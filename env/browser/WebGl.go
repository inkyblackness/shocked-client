package browser

import (
	"github.com/gopherjs/webgl"
)

// WebGl is a wrapper for the WebGL related operations.
type WebGl struct {
	gl *webgl.Context

	buffers           ObjectMapper
	programs          ObjectMapper
	shaders           ObjectMapper
	uniforms          ObjectMapper
	uniformsByProgram map[uint][]uint
}

// NewWebGl returns a new instance of WebGl, wrapping the provided context.
func NewWebGl(gl *webgl.Context) *WebGl {
	result := &WebGl{
		gl:                gl,
		buffers:           NewObjectMapper(),
		programs:          NewObjectMapper(),
		shaders:           NewObjectMapper(),
		uniforms:          NewObjectMapper(),
		uniformsByProgram: make(map[uint][]uint)}

	return result
}

// GetError implements the opengl.OpenGl interface.
func (web *WebGl) GetError() uint32 {
	return uint32(web.gl.GetError())
}

// Clear implements the opengl.OpenGl interface.
func (web *WebGl) Clear(mask uint32) {
	web.gl.Clear(int(mask))
}

// ClearColor implements the opengl.OpenGl interface.
func (web *WebGl) ClearColor(red float32, green float32, blue float32, alpha float32) {
	web.gl.ClearColor(red, green, blue, alpha)
}

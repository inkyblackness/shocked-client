package browser

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/webgl"
)

// WebGl is a wrapper for the WebGL related operations.
type WebGl struct {
	gl *webgl.Context

	buffers           ObjectMapper
	programs          ObjectMapper
	shaders           ObjectMapper
	uniforms          ObjectMapper
	uniformsByProgram map[uint32][]uint32
}

// NewWebGl returns a new instance of WebGl, wrapping the provided context.
func NewWebGl(gl *webgl.Context) *WebGl {
	result := &WebGl{
		gl:                gl,
		buffers:           NewObjectMapper(),
		programs:          NewObjectMapper(),
		shaders:           NewObjectMapper(),
		uniforms:          NewObjectMapper(),
		uniformsByProgram: make(map[uint32][]uint32)}

	return result
}

// AttachShader implements the opengl.OpenGl interface.
func (web *WebGl) AttachShader(program uint32, shader uint32) {
	objShader := web.shaders.get(shader)
	objProgram := web.programs.get(program)

	web.gl.AttachShader(objProgram, objShader)
}

// BindAttribLocation implements the opengl.OpenGl interface.
func (web *WebGl) BindAttribLocation(program uint32, index uint32, name string) {
	web.gl.BindAttribLocation(web.programs.get(program), int(index), name)
}

// BindBuffer implements the opengl.OpenGl interface.
func (web *WebGl) BindBuffer(target uint32, buffer uint32) {
	web.gl.BindBuffer(int(target), web.buffers.get(buffer))
}

// BufferData implements the opengl.OpenGl interface.
func (web *WebGl) BufferData(target uint32, size int, data interface{}, usage uint32) {
	web.gl.BufferData(int(target), data, int(usage))
}

// Clear implements the opengl.OpenGl interface.
func (web *WebGl) Clear(mask uint32) {
	web.gl.Clear(int(mask))
}

// ClearColor implements the opengl.OpenGl interface.
func (web *WebGl) ClearColor(red float32, green float32, blue float32, alpha float32) {
	web.gl.ClearColor(red, green, blue, alpha)
}

// CompileShader implements the opengl.OpenGl interface.
func (web *WebGl) CompileShader(shader uint32) {
	web.gl.CompileShader(web.shaders.get(shader))
}

// CreateProgram implements the opengl.OpenGl interface.
func (web *WebGl) CreateProgram() uint32 {
	key := web.programs.put(web.gl.CreateProgram())
	web.uniformsByProgram[key] = make([]uint32, 0)

	return key
}

// CreateShader implements the opengl.OpenGl interface.
func (web *WebGl) CreateShader(shaderType uint32) uint32 {
	return web.shaders.put(web.gl.CreateShader(int(shaderType)))
}

// DeleteBuffers implements the opengl.OpenGl interface.
func (web *WebGl) DeleteBuffers(buffers []uint32) {
	for _, buffer := range buffers {
		web.gl.DeleteBuffer(web.buffers.del(buffer))
	}
}

// DeleteProgram implements the opengl.OpenGl interface.
func (web *WebGl) DeleteProgram(program uint32) {
	web.gl.DeleteProgram(web.programs.del(program))
	for _, value := range web.uniformsByProgram[program] {
		web.uniforms.del(value)
	}
	delete(web.uniformsByProgram, program)
}

// DeleteShader implements the opengl.OpenGl interface.
func (web *WebGl) DeleteShader(shader uint32) {
	web.gl.DeleteShader(web.shaders.del(shader))
}

// DrawArrays implements the opengl.OpenGl interface.
func (web *WebGl) DrawArrays(mode uint32, first int32, count int32) {
	web.gl.DrawArrays(int(mode), int(first), int(count))
}

// Enable implements the opengl.OpenGl interface.
func (web *WebGl) Enable(cap uint32) {
	web.gl.Enable(int(cap))
}

// EnableVertexAttribArray implements the opengl.OpenGl interface.
func (web *WebGl) EnableVertexAttribArray(index uint32) {
	web.gl.EnableVertexAttribArray(int(index))
}

// GenBuffers implements the opengl.OpenGl interface.
func (web *WebGl) GenBuffers(n int32) []uint32 {
	ids := make([]uint32, n)

	for i := int32(0); i < n; i++ {
		ids[i] = web.buffers.put(web.gl.CreateBuffer())
	}

	return ids
}

// GetAttribLocation implements the opengl.OpenGl interface.
func (web *WebGl) GetAttribLocation(program uint32, name string) int32 {
	return int32(web.gl.GetAttribLocation(web.programs.get(program), name))
}

// GetError implements the opengl.OpenGl interface.
func (web *WebGl) GetError() uint32 {
	return uint32(web.gl.GetError())
}

// GetShaderInfoLog implements the opengl.OpenGl interface.
func (web *WebGl) GetShaderInfoLog(shader uint32) string {
	return web.gl.GetShaderInfoLog(web.shaders.get(shader))
}

func paramToInt(value *js.Object) int32 {
	result := int32(value.Int())

	if value.String() == "true" {
		result = 1
	}

	return result
}

// GetShaderParameter implements the opengl.OpenGl interface.
func (web *WebGl) GetShaderParameter(shader uint32, param uint32) int32 {
	value := web.gl.GetShaderParameter(web.shaders.get(shader), int(param))

	return paramToInt(value)
}

// GetProgramParameter implements the opengl.OpenGl interface.
func (web *WebGl) GetProgramParameter(program uint32, param uint32) int32 {
	value := web.gl.GetProgramParameteri(web.programs.get(program), int(param))
	//value := web.gl.Call("getProgramParameter", web.programs.get(program), int(param))

	return int32(value)
}

// GetUniformLocation implements the opengl.OpenGl interface.
func (web *WebGl) GetUniformLocation(program uint32, name string) int32 {
	uniform := web.gl.GetUniformLocation(web.programs.get(program), name)
	key := web.uniforms.put(uniform)

	web.uniformsByProgram[program] = append(web.uniformsByProgram[program], key)

	return int32(key)
}

// LinkProgram implements the opengl.OpenGl interface.
func (web *WebGl) LinkProgram(program uint32) {
	web.gl.LinkProgram(web.programs.get(program))
}

// ReadPixels implements the opengl.OpenGl interface.
func (web *WebGl) ReadPixels(x int32, y int32, width int32, height int32, format uint32, pixelType uint32, pixels interface{}) {
	//web.gl.ReadPixels(int(x), int(y), int(width), int(height), int(format), int(pixelType), pixels)
	//web.gl.Call("readPixels", x, y, width, height, int(format), int(pixelType), pixels)
}

// ShaderSource implements the opengl.OpenGl interface.
func (web *WebGl) ShaderSource(shader uint32, source string) {
	web.gl.ShaderSource(web.shaders.get(shader), source)
}

// UniformMatrix4fv implements the opengl.OpenGl interface.
func (web *WebGl) UniformMatrix4fv(location int32, transpose bool, value *[16]float32) {
	web.gl.UniformMatrix4fv(web.uniforms.get(uint32(location)), transpose, (*value)[:])
	//web.gl.Call("uniformMatrix4fv", web.uniforms.get(uint32(location)), transpose, *value)
}

// UseProgram implements the opengl.OpenGl interface.
func (web *WebGl) UseProgram(program uint32) {
	web.gl.UseProgram(web.programs.get(program))
}

// VertexAttribOffset implements the opengl.OpenGl interface.
func (web *WebGl) VertexAttribOffset(index uint32, size int32, attribType uint32, normalized bool, stride int32, offset int) {
	web.gl.VertexAttribPointer(int(index), int(size), int(attribType), normalized, int(stride), offset)
}

// Viewport implements the opengl.OpenGl interface.
func (web *WebGl) Viewport(x int32, y int32, width int32, height int32) {
	web.gl.Viewport(int(x), int(y), int(width), int(height))
}

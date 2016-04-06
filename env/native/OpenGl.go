package native

import (
	"strings"

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

// AttachShader implements the opengl.OpenGl interface.
func (native *OpenGl) AttachShader(program uint32, shader uint32) {
	gl.AttachShader(program, shader)
}

// BindAttribLocation implements the opengl.OpenGl interface.
func (native *OpenGl) BindAttribLocation(program uint32, index uint32, name string) {
	gl.BindAttribLocation(program, index, gl.Str(name+"\x00"))
}

// BindBuffer implements the opengl.OpenGl interface.
func (native *OpenGl) BindBuffer(target uint32, buffer uint32) {
	gl.BindBuffer(target, buffer)
}

// BindVertexArray implements the opengl.OpenGl interface.
func (native *OpenGl) BindVertexArray(array uint32) {
	gl.BindVertexArray(array)
}

// BufferData implements the opengl.OpenGl interface.
func (native *OpenGl) BufferData(target uint32, size int, data interface{}, usage uint32) {
	gl.BufferData(target, size, gl.Ptr(data), usage)
}

// Clear implements the opengl.OpenGl interface.
func (native *OpenGl) Clear(mask uint32) {
	gl.Clear(mask)
}

// ClearColor implements the opengl.OpenGl interface.
func (native *OpenGl) ClearColor(red float32, green float32, blue float32, alpha float32) {
	gl.ClearColor(red, green, blue, alpha)
}

// CompileShader implements the opengl.OpenGl interface.
func (native *OpenGl) CompileShader(shader uint32) {
	gl.CompileShader(shader)
}

// CreateProgram implements the opengl.OpenGl interface.
func (native *OpenGl) CreateProgram() uint32 {
	return gl.CreateProgram()
}

// CreateShader implements the opengl.OpenGl interface.
func (native *OpenGl) CreateShader(shaderType uint32) uint32 {
	return gl.CreateShader(shaderType)
}

// DeleteBuffers implements the opengl.OpenGl interface.
func (native *OpenGl) DeleteBuffers(buffers []uint32) {
	gl.DeleteBuffers(int32(len(buffers)), (*uint32)(&buffers[0]))
}

// DeleteProgram implements the opengl.OpenGl interface.
func (native *OpenGl) DeleteProgram(program uint32) {
	gl.DeleteProgram(program)
}

// DeleteShader implements the opengl.OpenGl interface.
func (native *OpenGl) DeleteShader(shader uint32) {
	gl.DeleteShader(shader)
}

// DeleteVertexArrays implements the opengl.OpenGl interface.
func (native *OpenGl) DeleteVertexArrays(arrays []uint32) {
	gl.DeleteVertexArrays(int32(len(arrays)), (*uint32)(&arrays[0]))
}

// DrawArrays implements the opengl.OpenGl interface.
func (native *OpenGl) DrawArrays(mode uint32, first int32, count int32) {
	gl.DrawArrays(mode, first, count)
}

// Enable implements the opengl.OpenGl interface.
func (native *OpenGl) Enable(cap uint32) {
	gl.Enable(cap)
}

// EnableVertexAttribArray implements the opengl.OpenGl interface.
func (native *OpenGl) EnableVertexAttribArray(index uint32) {
	gl.EnableVertexAttribArray(index)
}

// GenBuffers implements the opengl.OpenGl interface.
func (native *OpenGl) GenBuffers(n int32) []uint32 {
	buffers := make([]uint32, n, n)
	gl.GenBuffers(n, &buffers[0])
	return buffers
}

// GenVertexArrays implements the opengl.OpenGl interface.
func (native *OpenGl) GenVertexArrays(n int32) []uint32 {
	ids := make([]uint32, n, n)
	gl.GenVertexArrays(n, &ids[0])
	return ids
}

// GetAttribLocation implements the opengl.OpenGl interface.
func (native *OpenGl) GetAttribLocation(program uint32, name string) int32 {
	return gl.GetAttribLocation(program, gl.Str(name+"\x00"))
}

// GetError implements the opengl.OpenGl interface.
func (native *OpenGl) GetError() uint32 {
	return gl.GetError()
}

// GetProgramInfoLog implements the opengl.OpenGl interface.
func (native *OpenGl) GetProgramInfoLog(program uint32) string {
	logLength := native.GetProgramParameter(program, gl.INFO_LOG_LENGTH)
	log := strings.Repeat("\x00", int(logLength+1))
	gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
	return log
}

// GetProgramParameter implements the opengl.OpenGl interface.
func (native *OpenGl) GetProgramParameter(program uint32, param uint32) int32 {
	result := int32(0)
	gl.GetProgramiv(program, param, &result)
	return result
}

// GetShaderInfoLog implements the opengl.OpenGl interface.
func (native *OpenGl) GetShaderInfoLog(shader uint32) string {
	logLength := native.GetShaderParameter(shader, gl.INFO_LOG_LENGTH)
	log := strings.Repeat("\x00", int(logLength+1))
	gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
	return log
}

// GetShaderParameter implements the opengl.OpenGl interface.
func (native *OpenGl) GetShaderParameter(shader uint32, param uint32) int32 {
	result := int32(0)
	gl.GetShaderiv(shader, param, &result)
	return result
}

// GetUniformLocation implements the opengl.OpenGl interface.
func (native *OpenGl) GetUniformLocation(program uint32, name string) int32 {
	return gl.GetUniformLocation(program, gl.Str(name+"\x00"))
}

// LinkProgram implements the opengl.OpenGl interface.
func (native *OpenGl) LinkProgram(program uint32) {
	gl.LinkProgram(program)
}

// ReadPixels implements the opengl.OpenGl interface.
func (native *OpenGl) ReadPixels(x int32, y int32, width int32, height int32, format uint32, pixelType uint32, pixels interface{}) {
	gl.ReadPixels(x, y, width, height, format, pixelType, gl.Ptr(pixels))
}

// ShaderSource implements the opengl.OpenGl interface.
func (native *OpenGl) ShaderSource(shader uint32, source string) {
	csources, free := gl.Strs(source)
	defer free()

	gl.ShaderSource(shader, 1, csources, nil)
}

// UniformMatrix4fv implements the opengl.OpenGl interface.
func (native *OpenGl) UniformMatrix4fv(location int32, transpose bool, value *[16]float32) {
	count := int32(1)
	gl.UniformMatrix4fv(location, count, transpose, &value[0])
}

// UseProgram implements the opengl.OpenGl interface.
func (native *OpenGl) UseProgram(program uint32) {
	gl.UseProgram(program)
}

// VertexAttribOffset implements the opengl.OpenGl interface.
func (native *OpenGl) VertexAttribOffset(index uint32, size int32, attribType uint32, normalized bool, stride int32, offset int) {
	gl.VertexAttribPointer(index, size, attribType, normalized, stride, gl.PtrOffset(offset))
}

// Viewport implements the opengl.OpenGl interface.
func (native *OpenGl) Viewport(x int32, y int32, width int32, height int32) {
	gl.Viewport(x, y, width, height)
}

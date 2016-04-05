package opengl

// OpenGl describes an Open GL interface usable for all environments of this
// application. It is the common subset of WebGL (= OpenGL ES 2) and an equivalent
// API on the desktop.
type OpenGl interface {
	AttachShader(program uint32, shader uint32)

	BindAttribLocation(program uint32, index uint32, name string)
	BindBuffer(target uint32, buffer uint32)
	BindVertexArray(array uint32)
	BufferData(target uint32, size int, data interface{}, usage uint32)

	Clear(mask uint32)
	ClearColor(red float32, green float32, blue float32, alpha float32)

	CompileShader(shader uint32)

	CreateProgram() uint32
	CreateShader(shaderType uint32) uint32

	DeleteBuffers(buffers []uint32)
	DeleteProgram(program uint32)
	DeleteShader(shader uint32)
	DeleteVertexArrays(arrays []uint32)

	DrawArrays(mode uint32, first int32, count int32)

	Enable(cap uint32)
	EnableVertexAttribArray(index uint32)

	GenBuffers(n int32) []uint32
	GenVertexArrays(n int32) []uint32

	GetAttribLocation(program uint32, name string) int32
	GetError() uint32
	GetShaderInfoLog(shader uint32) string
	GetShaderParameter(shader uint32, param uint32) int32
	GetProgramInfoLog(program uint32) string
	GetProgramParameter(program uint32, param uint32) int32
	GetUniformLocation(program uint32, name string) int32

	LinkProgram(program uint32)

	ReadPixels(x int32, y int32, width int32, height int32, format uint32, pixelType uint32, pixels interface{})

	ShaderSource(shader uint32, source string)

	UniformMatrix4fv(location int32, transpose bool, value *[16]float32)
	UseProgram(program uint32)

	VertexAttribOffset(index uint32, size int32, attribType uint32, normalized bool, stride int32, offset int)
	Viewport(x int32, y int32, width int32, height int32)
}

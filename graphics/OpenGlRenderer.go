package graphics

import (
	"image/color"

	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/opengl"
)

var fillRectVertexShaderSource = `
  attribute vec3 vertexPosition;

  uniform mat4 projectionMatrix;

  void main(void) {
    gl_Position = projectionMatrix * vec4(vertexPosition, 1.0);
  }
`

var fillRectFragmentShaderSource = `
  #ifdef GL_ES
    precision mediump float;
  #endif

  uniform vec4 color;

  void main(void) {
    gl_FragColor = color;
  }
`

// OpenGlRenderer implements the ui.Renderer interface to provide rendering
// primitives using OpenGL.
type OpenGlRenderer struct {
	gl               opengl.OpenGl
	projectionMatrix *mgl.Mat4

	fillRectProgram         uint32
	fillRectVao             uint32
	vertexPositionBuffer    uint32
	vertexPositionAttrib    int32
	projectionMatrixUniform int32
	colorUniform            int32
}

// NewOpenGlRenderer returns a new instance of an OpenGlRenderer type.
func NewOpenGlRenderer(gl opengl.OpenGl, projectionMatrix *mgl.Mat4) *OpenGlRenderer {
	vertexShader, _ := opengl.CompileNewShader(gl, opengl.VERTEX_SHADER, fillRectVertexShaderSource)
	defer gl.DeleteShader(vertexShader)
	fragmentShader, _ := opengl.CompileNewShader(gl, opengl.FRAGMENT_SHADER, fillRectFragmentShaderSource)
	defer gl.DeleteShader(fragmentShader)
	fillRectProgram, _ := opengl.LinkNewProgram(gl, vertexShader, fragmentShader)

	renderer := &OpenGlRenderer{
		gl:               gl,
		projectionMatrix: projectionMatrix,

		fillRectProgram:         fillRectProgram,
		fillRectVao:             gl.GenVertexArrays(1)[0],
		vertexPositionBuffer:    gl.GenBuffers(1)[0],
		vertexPositionAttrib:    gl.GetAttribLocation(fillRectProgram, "vertexPosition"),
		projectionMatrixUniform: gl.GetUniformLocation(fillRectProgram, "projectionMatrix"),
		colorUniform:            gl.GetUniformLocation(fillRectProgram, "color")}

	return renderer
}

// Dispose clears any open resources.
func (renderer *OpenGlRenderer) Dispose() {

}

// FillRectangle implements the ui.Renderer interface
func (renderer *OpenGlRenderer) FillRectangle(left, top, right, bottom float32, fillColor color.Color) {
	renderer.withShader(func() {
		gl := renderer.gl

		var vertices = []float32{
			left, top, 0.0,
			right, top, 0.0,
			left, bottom, 0.0,

			left, bottom, 0.0,
			right, top, 0.0,
			right, bottom, 0.0}
		gl.BindBuffer(opengl.ARRAY_BUFFER, renderer.vertexPositionBuffer)
		gl.BufferData(opengl.ARRAY_BUFFER, len(vertices)*4, vertices, opengl.STATIC_DRAW)
		gl.BindBuffer(opengl.ARRAY_BUFFER, 0)

		gl.EnableVertexAttribArray(uint32(renderer.vertexPositionAttrib))
		gl.BindBuffer(opengl.ARRAY_BUFFER, renderer.vertexPositionBuffer)
		gl.VertexAttribOffset(uint32(renderer.vertexPositionAttrib), 3, opengl.FLOAT, false, 0, 0)
		gl.BindBuffer(opengl.ARRAY_BUFFER, 0)

		renderer.setMatrix(renderer.projectionMatrixUniform, renderer.projectionMatrix)
		renderer.setColor(renderer.colorUniform, fillColor)

		gl.DrawArrays(opengl.TRIANGLES, 0, 6)
	})
}

func (renderer *OpenGlRenderer) withShader(task func()) {
	gl := renderer.gl

	gl.UseProgram(renderer.fillRectProgram)
	gl.BindVertexArray(renderer.fillRectVao)

	defer func() {
		gl.BindVertexArray(0)
		gl.UseProgram(0)
	}()

	task()
}

func (renderer *OpenGlRenderer) setMatrix(uniform int32, matrix *mgl.Mat4) {
	matrixArray := ([16]float32)(*matrix)
	renderer.gl.UniformMatrix4fv(uniform, false, &matrixArray)
}

func (renderer *OpenGlRenderer) setColor(uniform int32, value color.Color) {
	r, g, b, a := value.RGBA()
	colorVector := [4]float32{float32(r) / 65535.0, float32(g) / 65535.0, float32(b) / 65535.0, float32(a) / 65535.0}

	renderer.gl.Uniform4fv(uniform, &colorVector)
}

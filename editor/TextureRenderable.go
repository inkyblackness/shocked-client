package editor

import (
	"fmt"
	"os"

	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/opengl"
)

var textureVertexShaderSource = `
  attribute vec3 vertexPosition;

  uniform mat4 modelMatrix;
  uniform mat4 viewMatrix;
  uniform mat4 projectionMatrix;

  varying vec2 position;

  void main(void) {
    gl_Position = projectionMatrix * viewMatrix * modelMatrix * vec4(vertexPosition, 1.0);

    position = vertexPosition.xy;
  }
`

var textureFragmentShaderSource = `
  #ifdef GL_ES
    precision mediump float;
  #endif

  uniform sampler2D palette;
  uniform sampler2D bitmap;

  varying vec2 position;

  void main(void) {
    vec4 pixel = texture2D(bitmap, position);
    vec4 color = texture2D(palette, vec2(pixel.a, 0.5));

    gl_FragColor = color;
  }
`

// TextureRenderable is a renderable for textures.
type TextureRenderable struct {
	gl opengl.OpenGl

	modelMatrix mgl.Mat4

	program                 uint32
	vertexArrayObject       uint32
	vertexPositionBuffer    uint32
	vertexPositionAttrib    int32
	modelMatrixUniform      int32
	viewMatrixUniform       int32
	projectionMatrixUniform int32

	paletteTexture uint32
	paletteUniform int32
	bitmapTexture  uint32
	bitmapUniform  int32
}

// NewTextureRenderable returns a new instance of a texture renderable
func NewTextureRenderable(gl opengl.OpenGl, width, height int, pixelData []byte,
	colorProvider func(index int) (byte, byte, byte, byte)) *TextureRenderable {
	vertexShader, err1 := opengl.CompileNewShader(gl, opengl.VERTEX_SHADER, textureVertexShaderSource)
	defer gl.DeleteShader(vertexShader)
	fragmentShader, err2 := opengl.CompileNewShader(gl, opengl.FRAGMENT_SHADER, textureFragmentShaderSource)
	defer gl.DeleteShader(fragmentShader)
	program, _ := opengl.LinkNewProgram(gl, vertexShader, fragmentShader)

	if err1 != nil {
		fmt.Fprintf(os.Stderr, "Failed to compile shader 1:\n", err1)
	}
	if err2 != nil {
		fmt.Fprintf(os.Stderr, "Failed to compile shader 2:\n", err2)
	}

	renderable := &TextureRenderable{
		gl:                      gl,
		program:                 program,
		modelMatrix:             mgl.Ident4().Mul4(mgl.Translate3D(64.0, 64.0, 0.0)).Mul4(mgl.Scale3D(128.0*5, 128.0*5, 1.0)),
		vertexArrayObject:       gl.GenVertexArrays(1)[0],
		vertexPositionBuffer:    gl.GenBuffers(1)[0],
		vertexPositionAttrib:    gl.GetAttribLocation(program, "vertexPosition"),
		modelMatrixUniform:      gl.GetUniformLocation(program, "modelMatrix"),
		viewMatrixUniform:       gl.GetUniformLocation(program, "viewMatrix"),
		projectionMatrixUniform: gl.GetUniformLocation(program, "projectionMatrix"),
		paletteTexture:          gl.GenTextures(1)[0],
		paletteUniform:          gl.GetUniformLocation(program, "palette"),
		bitmapTexture:           gl.GenTextures(1)[0],
		bitmapUniform:           gl.GetUniformLocation(program, "bitmap")}

	renderable.withShader(func() {
		gl.BindBuffer(opengl.ARRAY_BUFFER, renderable.vertexPositionBuffer)
		limit := float32(1.0)
		var vertices = []float32{
			0.0, 0.0, 0.0,
			limit, 0.0, 0.0,
			limit, limit, 0.0,

			limit, limit, 0.0,
			0.0, limit, 0.0,
			0.0, 0.0, 0.0}
		gl.BufferData(opengl.ARRAY_BUFFER, len(vertices)*4, vertices, opengl.STATIC_DRAW)

		gl.ActiveTexture(opengl.TEXTURE0 + 0)
		gl.BindTexture(opengl.TEXTURE_2D, renderable.paletteTexture)
		var palette [256 * 4]byte

		for i := 0; i < 256; i++ {
			r, g, b, a := colorProvider(i)
			palette[i*4+0] = r
			palette[i*4+1] = g
			palette[i*4+2] = b
			palette[i*4+3] = a
		}

		gl.TexImage2D(opengl.TEXTURE_2D, 0, opengl.RGBA, 256, 1, 0, opengl.RGBA, opengl.UNSIGNED_BYTE, palette)
		gl.TexParameteri(opengl.TEXTURE_2D, opengl.TEXTURE_MAG_FILTER, opengl.NEAREST)
		gl.TexParameteri(opengl.TEXTURE_2D, opengl.TEXTURE_MIN_FILTER, opengl.NEAREST)
		gl.GenerateMipmap(opengl.TEXTURE_2D)
		gl.BindTexture(opengl.TEXTURE_2D, 0)

		gl.ActiveTexture(opengl.TEXTURE0 + 1)
		gl.BindTexture(opengl.TEXTURE_2D, renderable.bitmapTexture)
		gl.TexImage2D(opengl.TEXTURE_2D, 0, opengl.ALPHA, int32(width), int32(height), 0, opengl.ALPHA, opengl.UNSIGNED_BYTE, pixelData)
		gl.TexParameteri(opengl.TEXTURE_2D, opengl.TEXTURE_MAG_FILTER, opengl.NEAREST)
		gl.TexParameteri(opengl.TEXTURE_2D, opengl.TEXTURE_MIN_FILTER, opengl.NEAREST)
		gl.GenerateMipmap(opengl.TEXTURE_2D)
		gl.BindTexture(opengl.TEXTURE_2D, 0)
	})

	return renderable
}

// Render renders
func (renderable *TextureRenderable) Render(context *RenderContext) {
	gl := renderable.gl

	renderable.withShader(func() {
		renderable.setMatrix(renderable.modelMatrixUniform, &renderable.modelMatrix)
		renderable.setMatrix(renderable.viewMatrixUniform, context.ViewMatrix())
		renderable.setMatrix(renderable.projectionMatrixUniform, context.ProjectionMatrix())

		gl.BindBuffer(opengl.ARRAY_BUFFER, renderable.vertexPositionBuffer)
		gl.VertexAttribOffset(uint32(renderable.vertexPositionAttrib), 3, opengl.FLOAT, false, 0, 0)

		textureUnit := int32(0)
		gl.ActiveTexture(opengl.TEXTURE0 + uint32(textureUnit))
		gl.BindTexture(opengl.TEXTURE_2D, renderable.paletteTexture)
		gl.Uniform1i(renderable.paletteUniform, textureUnit)

		textureUnit = 1
		gl.ActiveTexture(opengl.TEXTURE0 + uint32(textureUnit))
		gl.BindTexture(opengl.TEXTURE_2D, renderable.bitmapTexture)
		gl.Uniform1i(renderable.bitmapUniform, textureUnit)

		gl.DrawArrays(opengl.TRIANGLES, 0, 6)

		gl.BindTexture(opengl.TEXTURE_2D, 0)
	})
}

func (renderable *TextureRenderable) withShader(task func()) {
	gl := renderable.gl

	gl.UseProgram(renderable.program)
	gl.BindVertexArray(renderable.vertexArrayObject)
	gl.EnableVertexAttribArray(uint32(renderable.vertexPositionAttrib))

	defer func() {
		gl.EnableVertexAttribArray(0)
		gl.BindVertexArray(0)
		gl.UseProgram(0)
	}()

	task()
}

func (renderable *TextureRenderable) setMatrix(uniform int32, matrix *mgl.Mat4) {
	matrixArray := ([16]float32)(*matrix)
	renderable.gl.UniformMatrix4fv(uniform, false, &matrixArray)
}

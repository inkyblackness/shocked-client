package display

import (
	"fmt"
	"os"

	mgl32 "github.com/go-gl/mathgl/mgl32"
	mgl "github.com/go-gl/mathgl/mgl64"

	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/opengl"
)

// TileTextureMapRenderable is a renderable for textures.
type TileTextureMapRenderable struct {
	gl opengl.OpenGl

	program                 uint32
	vertexArrayObject       uint32
	vertexPositionBuffer    uint32
	vertexPositionAttrib    int32
	modelMatrixUniform      int32
	viewMatrixUniform       int32
	projectionMatrixUniform int32

	paletteUniform int32
	bitmapUniform  int32

	paletteTexture graphics.Texture

	tiles [][]graphics.Texture
}

// NewTileTextureMapRenderable returns a new instance of a renderable for tile maps
func NewTileTextureMapRenderable(gl opengl.OpenGl, paletteTexture graphics.Texture) *TileTextureMapRenderable {
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

	renderable := &TileTextureMapRenderable{
		gl:                      gl,
		program:                 program,
		vertexArrayObject:       gl.GenVertexArrays(1)[0],
		vertexPositionBuffer:    gl.GenBuffers(1)[0],
		vertexPositionAttrib:    gl.GetAttribLocation(program, "vertexPosition"),
		modelMatrixUniform:      gl.GetUniformLocation(program, "modelMatrix"),
		viewMatrixUniform:       gl.GetUniformLocation(program, "viewMatrix"),
		projectionMatrixUniform: gl.GetUniformLocation(program, "projectionMatrix"),
		paletteTexture:          paletteTexture,
		paletteUniform:          gl.GetUniformLocation(program, "palette"),
		bitmapUniform:           gl.GetUniformLocation(program, "bitmap"),
		tiles:                   make([][]graphics.Texture, 64)}

	for i := 0; i < 64; i++ {
		renderable.tiles[i] = make([]graphics.Texture, 64)
	}

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
	})

	return renderable
}

// Dispose releases any internal resources
func (renderable *TileTextureMapRenderable) Dispose() {
	renderable.gl.DeleteProgram(renderable.program)
	renderable.gl.DeleteBuffers([]uint32{renderable.vertexPositionBuffer})
	renderable.gl.DeleteVertexArrays([]uint32{renderable.vertexArrayObject})
}

// SetTileTexture sets the texture for the specified tile coordinate.
func (renderable *TileTextureMapRenderable) SetTileTexture(x, y int, tex graphics.Texture) {
	renderable.tiles[y][x] = tex
}

// Clear resets all tiles.
func (renderable *TileTextureMapRenderable) Clear() {
	for _, row := range renderable.tiles {
		for index := 0; index < len(row); index++ {
			row[index] = nil
		}
	}
}

// Render renders
func (renderable *TileTextureMapRenderable) Render(context *RenderContext) {
	gl := renderable.gl

	renderable.withShader(func() {
		renderable.setMatrix32(renderable.viewMatrixUniform, context.ViewMatrix())
		renderable.setMatrix32(renderable.projectionMatrixUniform, context.ProjectionMatrix())

		gl.BindBuffer(opengl.ARRAY_BUFFER, renderable.vertexPositionBuffer)
		gl.VertexAttribOffset(uint32(renderable.vertexPositionAttrib), 3, opengl.FLOAT, false, 0, 0)

		textureUnit := int32(0)
		gl.ActiveTexture(opengl.TEXTURE0 + uint32(textureUnit))
		gl.BindTexture(opengl.TEXTURE_2D, renderable.paletteTexture.Handle())
		gl.Uniform1i(renderable.paletteUniform, textureUnit)

		textureUnit = 1
		gl.ActiveTexture(opengl.TEXTURE0 + uint32(textureUnit))
		/*
			modelMatrix := mgl.Ident4().
				Mul4(mgl.Translate3D(float64(0)*32.0, float64(0)*32.0, 0.0)).
				Mul4(mgl.Scale3D(32.0, 32.0, 1.0))

			renderable.setMatrix64(renderable.modelMatrixUniform, &modelMatrix)
			/**/
		scaling := mgl.Scale3D(32.0, 32.0, 1.0)
		for y, row := range renderable.tiles {
			for x, tile := range row {
				if tile != nil {
					/* at zero
					modelMatrix := mgl.Ident4().
						Mul4(mgl.Translate3D(float64(0)*32.0, float64(0)*32.0, 0.0)).
						Mul4(mgl.Scale3D(32.0, 32.0, 1.0))

					renderable.setMatrix64(renderable.modelMatrixUniform, &modelMatrix)
					*/
					/**/
					modelMatrix := mgl.Translate3D(float64(x)*32.0, float64(y)*32.0, 0.0).
						Mul4(scaling)

					renderable.setMatrix64(renderable.modelMatrixUniform, &modelMatrix)
					/**/
					gl.BindTexture(opengl.TEXTURE_2D, tile.Handle())
					gl.Uniform1i(renderable.bitmapUniform, textureUnit)

					gl.DrawArrays(opengl.TRIANGLES, 0, 6)
				}
			}
		}

		gl.BindTexture(opengl.TEXTURE_2D, 0)
	})
}

func (renderable *TileTextureMapRenderable) withShader(task func()) {
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

func (renderable *TileTextureMapRenderable) setMatrix32(uniform int32, matrix *mgl32.Mat4) {
	matrixArray := ([16]float32)(*matrix)
	renderable.gl.UniformMatrix4fv(uniform, false, &matrixArray)
}

func (renderable *TileTextureMapRenderable) setMatrix64(uniform int32, matrix *mgl.Mat4) {
	var matrixArray [16]float32

	for index, value := range matrix {
		matrixArray[index] = float32(value)
	}
	renderable.gl.UniformMatrix4fv(uniform, false, &matrixArray)
}

package editor

import (
	"fmt"
	"os"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/opengl"
)

var vertexShaderSource = `
  attribute vec3 aVertexPosition;
  attribute vec4 aVertexColor;
  uniform mat4 uMVMatrix;
  uniform mat4 uPMatrix;
  varying vec4 vColor;
  void main(void) {
    gl_Position = uPMatrix * uMVMatrix * vec4(aVertexPosition, 1.0);
    vColor = aVertexColor;
  }
`

var fragmentShaderSource = `
  #ifdef GL_ES
    precision mediump float;
  #endif
  varying vec4 vColor;
  void main(void) {
    gl_FragColor = vColor;
  }
`

// MainApplication represents the core intelligence of the editor.
type MainApplication struct {
	glWindow env.OpenGlWindow
	gl       opengl.OpenGl

	vertexArrayObject            uint32
	vertexPosition               int32
	triangleVertexPositionBuffer uint32
	vertexColor                  int32
	triangleVertexColorBuffer    uint32

	pMatrix         mgl32.Mat4
	pMatrixUniform  int32
	mvMatrix        mgl32.Mat4
	mvMatrixUniform int32
}

// NewMainApplication returns a new instance of MainApplication.
func NewMainApplication() *MainApplication {
	return &MainApplication{}
}

// Init implements the env.Application interface.
func (app *MainApplication) Init(glWindow env.OpenGlWindow) {
	app.glWindow = glWindow

	glWindow.OnRender(app.render)

	builder := opengl.NewDebugBuilder(app.glWindow.OpenGl())

	/*
		builder.OnEntry(func(name string, param ...interface{}) {
			fmt.Fprintf(os.Stderr, "GL: [%-20s] %v ", name, param)
		})
		builder.OnExit(func(name string, result ...interface{}) {
			fmt.Fprintf(os.Stderr, "-> %v\n", result)
		})
	*/
	builder.OnError(func(name string, errorCodes []uint32) {
		errorStrings := make([]string, len(errorCodes))
		for index, errorCode := range errorCodes {
			errorStrings[index] = opengl.ErrorString(errorCode)
		}
		fmt.Fprintf(os.Stderr, "!!: [%-20s] %v -> %v\n", name, errorCodes, errorStrings)
	})

	app.gl = builder.Build()

	app.initShaders()
	app.initBuffers()

	app.gl.ClearColor(0.0, 0.0, 0.0, 1.0)
}

func (app *MainApplication) render() {
	gl := app.gl
	width, height := app.glWindow.Size()

	//fmt.Fprintf(os.Stderr, "Size: %vx%v\n", width, height)

	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(opengl.COLOR_BUFFER_BIT | opengl.DEPTH_BUFFER_BIT)

	app.pMatrix = mgl32.Ortho2D(0, float32(width), float32(height), 0)
	app.mvMatrix = mgl32.Ident4().Mul4(mgl32.Translate3D(20, 20, 0.0))
	app.setMatrixUniforms()

	gl.BindBuffer(opengl.ARRAY_BUFFER, app.triangleVertexPositionBuffer)
	gl.VertexAttribOffset(uint32(app.vertexPosition), 3, opengl.FLOAT, false, 0, 0)

	gl.BindBuffer(opengl.ARRAY_BUFFER, app.triangleVertexColorBuffer)
	gl.VertexAttribOffset(uint32(app.vertexColor), 4, opengl.FLOAT, false, 0, 0)

	gl.DrawArrays(opengl.TRIANGLES, 0, 3)
}

func (app *MainApplication) initShaders() {
	gl := app.gl
	fragmentShader, err1 := opengl.CompileNewShader(gl, opengl.FRAGMENT_SHADER, fragmentShaderSource)
	vertexShader, err2 := opengl.CompileNewShader(gl, opengl.VERTEX_SHADER, vertexShaderSource)
	if err1 != nil {
		fmt.Fprintf(os.Stderr, "Failed to compile shader 1!\n%v\n", err1)
	}
	if err2 != nil {
		fmt.Fprintf(os.Stderr, "Failed to compile shader 2!\n%v\n", err2)
	}

	program, err3 := opengl.LinkNewProgram(gl, fragmentShader, vertexShader)
	if err3 != nil {
		fmt.Fprintf(os.Stderr, "Failed to link program!\n%v\n", err3)
	}

	gl.UseProgram(program)

	app.vertexArrayObject = gl.GenVertexArrays(1)[0]
	gl.BindVertexArray(app.vertexArrayObject)

	app.vertexPosition = gl.GetAttribLocation(program, "aVertexPosition")
	gl.EnableVertexAttribArray(uint32(app.vertexPosition))
	app.vertexColor = gl.GetAttribLocation(program, "aVertexColor")
	gl.EnableVertexAttribArray(uint32(app.vertexColor))

	app.pMatrixUniform = gl.GetUniformLocation(program, "uPMatrix")
	app.mvMatrixUniform = gl.GetUniformLocation(program, "uMVMatrix")
}

func (app *MainApplication) initBuffers() {
	gl := app.gl

	app.triangleVertexPositionBuffer = gl.GenBuffers(1)[0]
	gl.BindBuffer(opengl.ARRAY_BUFFER, app.triangleVertexPositionBuffer)
	var vertices = []float32{
		0.0, 0.0, 0.0,
		-10.0, 10.0, 0.0,
		10.0, 10.0, 0.0}
	gl.BufferData(opengl.ARRAY_BUFFER, len(vertices)*4, vertices, opengl.STATIC_DRAW)

	app.triangleVertexColorBuffer = gl.GenBuffers(1)[0]
	gl.BindBuffer(opengl.ARRAY_BUFFER, app.triangleVertexColorBuffer)
	var colors = []float32{
		1.0, 0.0, 0.0, 1.0,
		0.0, 1.0, 0.0, 1.0,
		0.0, 0.0, 1.0, 1.0}
	gl.BufferData(opengl.ARRAY_BUFFER, len(colors)*4, colors, opengl.STATIC_DRAW)
}

func (app *MainApplication) setMatrixUniforms() {
	gl := app.gl
	pMatrixArr := ([16]float32)(app.pMatrix)
	gl.UniformMatrix4fv(app.pMatrixUniform, false, &pMatrixArr)
	mvMatrixArr := ([16]float32)(app.mvMatrix)
	gl.UniformMatrix4fv(app.mvMatrixUniform, false, &mvMatrixArr)
}

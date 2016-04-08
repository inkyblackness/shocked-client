package editor

import (
	"fmt"
	"math"
	"os"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/opengl"
)

// MainApplication represents the core intelligence of the editor.
type MainApplication struct {
	glWindow env.OpenGlWindow
	gl       opengl.OpenGl

	requestedZoomLevel float32

	gridRenderable *GridRenderable
}

// NewMainApplication returns a new instance of MainApplication.
func NewMainApplication() *MainApplication {
	return &MainApplication{
		requestedZoomLevel: 0}
}

// Init implements the env.Application interface.
func (app *MainApplication) Init(glWindow env.OpenGlWindow) {
	app.glWindow = glWindow

	glWindow.OnRender(app.render)
	glWindow.OnMouseMove(app.onMouseMove)
	glWindow.OnMouseButtonDown(app.onMouseButtonDown)
	glWindow.OnMouseButtonUp(app.onMouseButtonUp)
	glWindow.OnMouseScroll(app.onMouseScroll)

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

	app.gl.Enable(opengl.BLEND)
	app.gl.BlendFunc(opengl.SRC_ALPHA, opengl.ONE_MINUS_SRC_COLOR)
	app.gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	app.gridRenderable = NewGridRenderable(app.gl)
}

func (app *MainApplication) render() {
	gl := app.gl
	width, height := app.glWindow.Size()
	scaleFactor := float32(math.Pow(2.0, float64(app.requestedZoomLevel)))

	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(opengl.COLOR_BUFFER_BIT | opengl.DEPTH_BUFFER_BIT)

	context := RenderContext{
		viewportWidth:    width,
		viewportHeight:   height,
		viewMatrix:       mgl32.Ident4().Mul4(mgl32.Scale3D(scaleFactor, scaleFactor, 1.0)),
		projectionMatrix: mgl32.Ortho2D(0, float32(width), float32(height), 0)}

	app.gridRenderable.Render(&context)
}

func (app *MainApplication) onMouseMove(x float32, y float32) {
	fmt.Fprintf(os.Stderr, "mouse: %v, %v\n", x, y)
}

func (app *MainApplication) onMouseButtonDown(mouseButton uint32) {
	fmt.Fprintf(os.Stderr, "down: 0x%08X\n", mouseButton)
}

func (app *MainApplication) onMouseButtonUp(mouseButton uint32) {
	fmt.Fprintf(os.Stderr, "up: 0x%08X\n", mouseButton)
}

func (app *MainApplication) onMouseScroll(dx float32, dy float32) {
	if dy > 0 {
		app.Zoom(-0.5)
	}
	if dy < 0 {
		app.Zoom(0.5)
	}
}

// Zoom adjusts the requested zoom level by given delta. Positive values zoom in.
func (app *MainApplication) Zoom(levelDelta float32) {
	newValue := app.requestedZoomLevel + levelDelta
	if newValue < ZoomLevelMin {
		newValue = ZoomLevelMin
	}
	if newValue > ZoomLevelMax {
		newValue = ZoomLevelMax
	}
	app.requestedZoomLevel = newValue
}

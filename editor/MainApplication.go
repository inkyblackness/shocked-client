package editor

import (
	"fmt"
	"os"

	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/editor/camera"
	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/opengl"
)

// MainApplication represents the core intelligence of the editor.
type MainApplication struct {
	glWindow env.OpenGlWindow
	gl       opengl.OpenGl

	mouseX, mouseY   float32
	mouseMoveCapture func()

	view *camera.LimitedCamera

	gridRenderable *GridRenderable
}

// NewMainApplication returns a new instance of MainApplication.
func NewMainApplication() *MainApplication {
	camLimit := (TilesPerMapSide - 1) * TileBaseLength

	return &MainApplication{
		mouseMoveCapture: func() {},
		view:             camera.NewLimited(ZoomLevelMin, ZoomLevelMax, 0, camLimit)}
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

	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(opengl.COLOR_BUFFER_BIT | opengl.DEPTH_BUFFER_BIT)

	context := RenderContext{
		viewportWidth:    width,
		viewportHeight:   height,
		viewMatrix:       app.view.ViewMatrix(),
		projectionMatrix: mgl.Ortho2D(0, float32(width), float32(height), 0)}

	app.gridRenderable.Render(&context)
}

func (app *MainApplication) unprojectPixel(pixelX, pixelY float32) (x, y float32) {
	pixelVec := mgl.Vec4{pixelX, pixelY, 0.0, 1.0}
	invertedView := app.view.ViewMatrix().Inv()
	result := invertedView.Mul4x1(pixelVec)

	return result[0], result[1]
}

func (app *MainApplication) onMouseMove(x float32, y float32) {
	app.mouseX, app.mouseY = x, y

	//worldMouseX, worldMouseY := app.unprojectPixel(app.mouseX, app.mouseY)
	//fmt.Fprintf(os.Stderr, "mv: pixelMouse: %v, %v; worldMouse: %v, %v\n", app.mouseX, app.mouseY, worldMouseX, worldMouseY)

	app.mouseMoveCapture()
}

func (app *MainApplication) onMouseButtonDown(mouseButton uint32) {
	if (mouseButton & env.MousePrimary) == env.MousePrimary {
		lastMouseX, lastMouseY := app.mouseX, app.mouseY

		app.mouseMoveCapture = func() {
			lastWorldMouseX, lastWorldMouseY := app.unprojectPixel(lastMouseX, lastMouseY)
			worldMouseX, worldMouseY := app.unprojectPixel(app.mouseX, app.mouseY)

			app.view.MoveBy(worldMouseX-lastWorldMouseX, worldMouseY-lastWorldMouseY)
			lastMouseX, lastMouseY = app.mouseX, app.mouseY
		}
	}
}

func (app *MainApplication) onMouseButtonUp(mouseButton uint32) {
	if (mouseButton & env.MousePrimary) == env.MousePrimary {
		app.mouseMoveCapture = func() {}
	}
}

func (app *MainApplication) onMouseScroll(dx float32, dy float32) {
	worldMouseX, worldMouseY := app.unprojectPixel(app.mouseX, app.mouseY)
	if dy > 0 {
		app.view.ZoomAt(-0.5, worldMouseX, worldMouseY)
	}
	if dy < 0 {
		app.view.ZoomAt(0.5, worldMouseX, worldMouseY)
	}
}

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

	mouseX, mouseY   float32
	mouseMoveCapture func()

	focusX, focusY float32

	requestedZoomLevel       float32
	viewOffsetX, viewOffsetY float32

	viewMatrix mgl32.Mat4

	gridRenderable *GridRenderable
}

// NewMainApplication returns a new instance of MainApplication.
func NewMainApplication() *MainApplication {
	return &MainApplication{
		requestedZoomLevel: 0,
		viewMatrix:         mgl32.Ident4(),
		mouseMoveCapture:   func() {}}
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

	app.updateViewMatrix()
}

func (app *MainApplication) render() {
	gl := app.gl
	width, height := app.glWindow.Size()

	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(opengl.COLOR_BUFFER_BIT | opengl.DEPTH_BUFFER_BIT)

	context := RenderContext{
		viewportWidth:    width,
		viewportHeight:   height,
		viewMatrix:       app.viewMatrix,
		projectionMatrix: mgl32.Ortho2D(0, float32(width), float32(height), 0)}

	app.gridRenderable.Render(&context)
}

func (app *MainApplication) unprojectPixel(pixelX, pixelY float32) (x, y float32) {
	pixelVec := mgl32.Vec4{pixelX, pixelY, 0.0, 1.0}
	invertedView := app.viewMatrix.Inv()
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

			app.ScrollBy(worldMouseX-lastWorldMouseX, worldMouseY-lastWorldMouseY)
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
		app.ZoomAt(-0.5, worldMouseX, worldMouseY)
	}
	if dy < 0 {
		app.ZoomAt(0.5, worldMouseX, worldMouseY)
	}
}

// ScrollBy adjusts the requested view offset by given delta values in world coordinates.
func (app *MainApplication) ScrollBy(dx, dy float32) {
	app.ScrollTo(app.viewOffsetX+dx, app.viewOffsetY+dy)
}

// ScrollTo sets the requested view offset to the given world coordinates.
func (app *MainApplication) ScrollTo(worldX, worldY float32) {
	limit := (TilesPerMapSide - 1) * TileBaseLength

	limitOffset := func(offset float32) float32 {
		result := offset
		if offset < -limit {
			result = -limit
		}
		if offset > 0 {
			result = 0
		}
		return result
	}

	app.viewOffsetX = limitOffset(worldX)
	app.viewOffsetY = limitOffset(worldY)
	app.updateViewMatrix()
}

// ZoomAt adjusts the requested zoom level by given delta, centered around give position.
// Positive values zoom in.
func (app *MainApplication) ZoomAt(levelDelta float32, x, y float32) {
	app.focusX, app.focusY = x, y
	app.Zoom(levelDelta)
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

	focusPoint := mgl32.Vec4{app.focusX, app.focusY, 0.0, 1.0}
	oldPixel := app.viewMatrix.Mul4x1(focusPoint)

	app.updateViewMatrix()

	newPixel := app.viewMatrix.Mul4x1(focusPoint)
	scaleFactor := app.scaleFactor()
	app.ScrollBy(-(newPixel[0]-oldPixel[0])/scaleFactor, -(newPixel[1]-oldPixel[1])/scaleFactor)
}

func (app *MainApplication) scaleFactor() float32 {
	return float32(math.Pow(2.0, float64(app.requestedZoomLevel)))
}

func (app *MainApplication) updateViewMatrix() {
	scaleFactor := app.scaleFactor()
	app.viewMatrix = mgl32.Ident4().
		Mul4(mgl32.Scale3D(scaleFactor, scaleFactor, 1.0)).
		Mul4(mgl32.Translate3D(app.viewOffsetX, app.viewOffsetY, 0))
}

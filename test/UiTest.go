package main

import (
	"fmt"
	"image/color"
	"os"
	//"runtime/pprof"

	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/env/native"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/opengl"
	"github.com/inkyblackness/shocked-client/ui"
)

func main() {
	deferrer := make(chan func(), 100)
	defer close(deferrer)

	/*
		f, err := os.Create("profile")
		if err != nil {
			fmt.Println(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	*/

	app := newUITestApplication()

	native.Run(app, deferrer)
}

type uiTestApplication struct {
	glWindow env.OpenGlWindow
	gl       opengl.OpenGl

	projectionMatrix mgl.Mat4

	rootArea   *ui.Area
	uiRenderer *graphics.OpenGlRenderer
}

func newUITestApplication() *uiTestApplication {
	return &uiTestApplication{}
}

func (app *uiTestApplication) Init(glWindow env.OpenGlWindow) {
	app.setWindow(glWindow)
	app.initOpenGl()
	app.setDebugOpenGl()

	app.uiRenderer = graphics.NewOpenGlRenderer(app.gl, &app.projectionMatrix)

	app.initInterface()

	app.onWindowResize(glWindow.Size())
}

func (app *uiTestApplication) setWindow(glWindow env.OpenGlWindow) {
	app.glWindow = glWindow
	app.gl = glWindow.OpenGl()

	glWindow.OnRender(app.render)
	glWindow.OnResize(app.onWindowResize)
}

func (app *uiTestApplication) initOpenGl() {
	app.gl.Disable(opengl.DEPTH_TEST)
	app.gl.Enable(opengl.BLEND)
	app.gl.BlendFunc(opengl.SRC_ALPHA, opengl.ONE_MINUS_SRC_ALPHA)
	app.gl.ClearColor(0.0, 0.0, 0.0, 1.0)
}

func (app *uiTestApplication) setDebugOpenGl() {
	builder := opengl.NewDebugBuilder(app.gl)

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
}

func (app *uiTestApplication) initInterface() {
	rootBuilder := ui.NewAreaBuilder()

	rootBuilder.SetRight(ui.NewAbsoluteAnchor(0.0))
	rootBuilder.SetBottom(ui.NewAbsoluteAnchor(0.0))
	rootBuilder.OnRender(func(area *ui.Area, renderer ui.Renderer) {
		renderer.FillRectangle(area.Left().Value(), area.Top().Value(), area.Right().Value(), area.Bottom().Value(),
			color.RGBA{0x20, 0x40, 0x70, 0xFF})
	})

	app.rootArea = rootBuilder.Build()

	//
	mainPanelBuilder := ui.NewAreaBuilder()
	mainPanelBuilder.SetParent(app.rootArea)

	mainPanelRight := ui.NewOffsetAnchor(app.rootArea.Right(), -5.0)
	mainPanelBuilder.SetRight(mainPanelRight)
	mainPanelBuilder.SetLeft(ui.NewOffsetAnchor(mainPanelRight, -20.0))
	mainPanelTop := ui.NewRelativeAnchor(app.rootArea.Top(), app.rootArea.Bottom(), 0.1)
	mainPanelBuilder.SetTop(mainPanelTop)
	mainPanelBuilder.SetBottom(app.rootArea.Bottom())
	mainPanelBuilder.OnRender(func(area *ui.Area, renderer ui.Renderer) {
		renderer.FillRectangle(area.Left().Value(), area.Top().Value(), area.Right().Value(), area.Bottom().Value(),
			color.RGBA{0x70, 0x10, 0x40, 0x80})
	})
	mainPanelBuilder.Build()

	//
	horizontalCenter := ui.NewRelativeAnchor(app.rootArea.Left(), app.rootArea.Right(), 0.5)
	verticalCenter := ui.NewRelativeAnchor(app.rootArea.Top(), app.rootArea.Bottom(), 0.5)

	minPanelWidth := float32(50.0)
	minPanelHeight := float32(30.0)

	centerPanelBuilder := ui.NewAreaBuilder()
	centerPanelBuilder.SetParent(app.rootArea)

	centerPanelBuilder.SetLeft(ui.NewOffsetAnchor(horizontalCenter, minPanelWidth/-2.0))
	centerPanelBuilder.SetRight(ui.NewOffsetAnchor(horizontalCenter, minPanelWidth/2.0))
	centerPanelBuilder.SetTop(ui.NewOffsetAnchor(verticalCenter, minPanelHeight/-2.0))
	centerPanelBuilder.SetBottom(ui.NewOffsetAnchor(verticalCenter, minPanelHeight/2.0))

	centerPanelBuilder.OnRender(func(area *ui.Area, renderer ui.Renderer) {
		renderer.FillRectangle(area.Left().Value(), area.Top().Value(), area.Right().Value(), area.Bottom().Value(),
			color.RGBA{0x40, 0x00, 0x40, 0xC0})
	})
	centerPanelBuilder.Build()

	//
	sidePanelBuilder := ui.NewAreaBuilder()
	sidePanelBuilder.SetParent(app.rootArea)

	sidePanelLeft := ui.NewOffsetAnchor(app.rootArea.Left(), 10.0)
	sidePanelBuilder.SetLeft(sidePanelLeft)
	sidePanelBuilder.SetTop(ui.NewOffsetAnchor(app.rootArea.Top(), 10.0))
	sidePanelBuilder.SetBottom(ui.NewOffsetAnchor(app.rootArea.Bottom(), -10.0))

	minRight := ui.NewOffsetAnchor(sidePanelLeft, 200.0)
	maxRight := ui.NewOffsetAnchor(sidePanelLeft, 400.0)
	sidePanelBuilder.SetRight(ui.NewLimitedAnchor(minRight, maxRight, ui.NewRelativeAnchor(app.rootArea.Left(), app.rootArea.Right(), 0.4)))

	sidePanelBuilder.OnRender(func(area *ui.Area, renderer ui.Renderer) {
		renderer.FillRectangle(area.Left().Value(), area.Top().Value(), area.Right().Value(), area.Bottom().Value(),
			color.RGBA{0x00, 0x60, 0x40, 0xC0})
	})
	sidePanelBuilder.Build()
}

func (app *uiTestApplication) onWindowResize(width int, height int) {
	app.projectionMatrix = mgl.Ortho2D(0.0, float32(width), float32(height), 0.0)
	app.gl.Viewport(0, 0, int32(width), int32(height))

	app.rootArea.Right().RequestValue(float32(width))
	app.rootArea.Bottom().RequestValue(float32(height))
}

func (app *uiTestApplication) render() {
	gl := app.gl

	gl.Clear(opengl.COLOR_BUFFER_BIT)
	app.rootArea.Render(app.uiRenderer)
}

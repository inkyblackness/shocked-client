package editor

import (
	"fmt"
	"os"
	"time"

	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/env/keys"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/opengl"
)

// MainApplication represents the core intelligence of the editor.
type MainApplication struct {
	lastElapsedTick time.Time
	elapsedMSec     int64

	store DataStore

	glWindow                  env.OpenGlWindow
	windowWidth, windowHeight float32
	gl                        opengl.OpenGl
	projectionMatrix          mgl.Mat4

	mouseX, mouseY   float32
	mouseDragged     bool
	mouseMoveCapture func()

	defaultFont graphics.TextRenderer
	defaultIcon *graphics.BitmapTexture
}

// NewMainApplication returns a new instance of MainApplication.
func NewMainApplication(store DataStore) *MainApplication {
	app := &MainApplication{
		projectionMatrix: mgl.Ident4(),
		lastElapsedTick:  time.Now(),
		store:            store,
		mouseMoveCapture: func() {}}

	return app
}

// Init implements the env.Application interface.
func (app *MainApplication) Init(glWindow env.OpenGlWindow) {
	app.setWindow(glWindow)
	app.setDebugOpenGl()
	app.initOpenGl()

	app.onWindowResize(glWindow.Size())
}

func (app *MainApplication) setWindow(glWindow env.OpenGlWindow) {
	app.glWindow = glWindow
	app.gl = glWindow.OpenGl()

	glWindow.OnRender(app.render)
	glWindow.OnResize(app.onWindowResize)
	glWindow.OnMouseMove(app.onMouseMove)
	glWindow.OnMouseButtonDown(app.onMouseButtonDown)
	glWindow.OnMouseButtonUp(app.onMouseButtonUp)
	glWindow.OnMouseScroll(app.onMouseScroll)
	glWindow.OnKey(app.onKey)
	glWindow.OnModifier(app.onModifier)
	glWindow.OnCharCallback(app.onChar)
}

func (app *MainApplication) setDebugOpenGl() {
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

func (app *MainApplication) initOpenGl() {
	app.gl.Disable(opengl.DEPTH_TEST)
	app.gl.Enable(opengl.BLEND)
	app.gl.BlendFunc(opengl.SRC_ALPHA, opengl.ONE_MINUS_SRC_ALPHA)
	app.gl.ClearColor(0.0, 0.0, 0.0, 1.0)
}

func (app *MainApplication) simpleStoreFailure(info string) FailureFunc {
	return func() {
		fmt.Fprintf(os.Stderr, "Failed to process store query <%s>\n", info)
	}
}

func (app *MainApplication) updateElapsedNano() {
	now := time.Now()
	diff := now.Sub(app.lastElapsedTick).Nanoseconds()

	if diff > 0 {
		app.elapsedMSec += diff / 1000000
	}
	app.lastElapsedTick = now
}

func (app *MainApplication) onWindowResize(width int, height int) {
	app.windowWidth, app.windowHeight = float32(width), float32(height)
	app.projectionMatrix = mgl.Ortho2D(0.0, app.windowWidth, app.windowHeight, 0.0)
	app.gl.Viewport(0, 0, int32(width), int32(height))
}

func (app *MainApplication) render() {
	gl := app.gl

	gl.Clear(opengl.COLOR_BUFFER_BIT)

	app.updateElapsedNano()

	//viewMatrix := mgl.Ident4()
	//context := display.NewBasicRenderContext(app.windowWidth, app.windowHeight, app.projectionMatrix, viewMatrix)
}

func (app *MainApplication) onMouseMove(x float32, y float32) {
	app.mouseX, app.mouseY = x, y

	app.mouseMoveCapture()
}

func (app *MainApplication) onMouseButtonDown(mouseButton uint32, modifier keys.Modifier) {
	app.mouseDragged = false
	app.mouseMoveCapture = func() {
		app.mouseDragged = true
	}
}

func (app *MainApplication) onMouseButtonUp(mouseButton uint32, modifier keys.Modifier) {
	if (mouseButton & env.MousePrimary) == env.MousePrimary {
		app.mouseMoveCapture = func() {}
		if !app.mouseDragged {
			app.onMouseClick(modifier)
		}
	}
}

func (app *MainApplication) onMouseScroll(dx float32, dy float32) {
}

func (app *MainApplication) onMouseClick(modifierMask keys.Modifier) {
}

func (app *MainApplication) onKey(key keys.Key, modifier keys.Modifier) {
	fmt.Printf("down: %v [%v]\n", key, modifier)
}

func (app *MainApplication) onModifier(modifier keys.Modifier) {
	fmt.Printf(" mod: [%v]\n", modifier)
}

func (app *MainApplication) onChar(char rune) {
	fmt.Printf("char: %v\n", string(char))
}

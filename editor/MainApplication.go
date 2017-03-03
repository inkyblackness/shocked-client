package editor

import (
	"fmt"
	"os"
	"time"

	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/editor/model"
	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/env/keys"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/opengl"
	"github.com/inkyblackness/shocked-client/ui"
	"github.com/inkyblackness/shocked-client/ui/events"
	dataModel "github.com/inkyblackness/shocked-model"
)

// MainApplication represents the core intelligence of the editor.
type MainApplication struct {
	lastElapsedTick time.Time
	elapsedMSec     int64

	store        dataModel.DataStore
	modelAdapter *model.Adapter

	glWindow                  env.OpenGlWindow
	windowWidth, windowHeight float32
	gl                        opengl.OpenGl
	projectionMatrix          mgl.Mat4

	mouseX, mouseY float32
	mouseButtons   uint32
	keyModifier    keys.Modifier

	rootArea           *ui.Area
	defaultFontPainter graphics.TextPainter
	uiTextPalette      *graphics.PaletteTexture
	rectRenderer       *graphics.RectangleRenderer
	uiTextRenderer     *graphics.BitmapTextureRenderer
}

// NewMainApplication returns a new instance of MainApplication.
func NewMainApplication(store dataModel.DataStore) *MainApplication {
	app := &MainApplication{
		projectionMatrix:   mgl.Ident4(),
		lastElapsedTick:    time.Now(),
		store:              store,
		modelAdapter:       model.NewAdapter(store),
		defaultFontPainter: graphics.NewBitmapTextPainter(defaultFont)}

	return app
}

// Init implements the env.Application interface.
func (app *MainApplication) Init(glWindow env.OpenGlWindow) {
	app.setWindow(glWindow)
	app.setDebugOpenGl()
	app.initOpenGl()

	app.initInterface()

	app.onWindowResize(glWindow.Size())

	app.modelAdapter.SetMessage("Ready.")
	app.modelAdapter.RequestProject("(inplace)")
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

func (app *MainApplication) initInterface() {
	app.rectRenderer = graphics.NewRectangleRenderer(app.gl, &app.projectionMatrix)

	uiTextPalette := [][4]byte{
		{0x00, 0x00, 0x00, 0x00},
		{0x80, 0x94, 0x54, 0xFF},
		{0x00, 0x00, 0x00, 0xC0}}
	app.uiTextPalette = graphics.NewPaletteTexture(app.gl, func(index int) (byte, byte, byte, byte) {
		fetchIndex := index
		if fetchIndex >= len(uiTextPalette) {
			fetchIndex = 0
		}
		entry := uiTextPalette[fetchIndex]
		return entry[0], entry[1], entry[2], entry[3]
	})
	viewMatrix := mgl.Ident4()
	uiRenderContext := graphics.NewBasicRenderContext(app.gl, &app.projectionMatrix, &viewMatrix)
	app.uiTextRenderer = graphics.NewBitmapTextureRenderer(uiRenderContext, app.uiTextPalette)

	app.rootArea = newRootArea(app)
}

func (app *MainApplication) updateElapsedNano() {
	now := time.Now()
	diff := now.Sub(app.lastElapsedTick).Nanoseconds()

	if diff > 0 {
		app.elapsedMSec += diff / time.Millisecond.Nanoseconds()
	}
	app.lastElapsedTick = now
}

func (app *MainApplication) onWindowResize(width int, height int) {
	app.projectionMatrix = mgl.Ortho2D(0.0, float32(width), float32(height), 0.0)
	app.gl.Viewport(0, 0, int32(width), int32(height))

	app.rootArea.Right().RequestValue(float32(width))
	app.rootArea.Bottom().RequestValue(float32(height))
}

func (app *MainApplication) render() {
	gl := app.gl

	gl.Clear(opengl.COLOR_BUFFER_BIT)

	app.updateElapsedNano()
	app.rootArea.Render()
}

func (app *MainApplication) onMouseMove(x float32, y float32) {
	app.mouseX, app.mouseY = x, y
	app.rootArea.DispatchPositionalEvent(events.NewMouseMoveEvent(
		app.mouseX, app.mouseY, uint32(app.keyModifier), app.mouseButtons))
}

func (app *MainApplication) onMouseButtonDown(mouseButton uint32, modifier keys.Modifier) {
	app.mouseButtons |= mouseButton
	app.rootArea.DispatchPositionalEvent(events.NewMouseButtonEvent(events.MouseButtonDownEventType,
		app.mouseX, app.mouseY, uint32(app.keyModifier), app.mouseButtons, mouseButton))
}

func (app *MainApplication) onMouseButtonUp(mouseButton uint32, modifier keys.Modifier) {
	app.mouseButtons &= ^mouseButton
	app.rootArea.DispatchPositionalEvent(events.NewMouseButtonEvent(events.MouseButtonUpEventType,
		app.mouseX, app.mouseY, uint32(app.keyModifier), app.mouseButtons, mouseButton))
}

func (app *MainApplication) onMouseScroll(dx float32, dy float32) {
	app.rootArea.DispatchPositionalEvent(events.NewMouseScrollEvent(
		app.mouseX, app.mouseY, uint32(app.keyModifier), app.mouseButtons, dx, dy))
}

func (app *MainApplication) onMouseClick(modifierMask keys.Modifier) {
}

func (app *MainApplication) onKey(key keys.Key, modifier keys.Modifier) {
	app.keyModifier = modifier
}

func (app *MainApplication) onModifier(modifier keys.Modifier) {
	app.keyModifier = modifier
}

func (app *MainApplication) onChar(char rune) {
}

// ModelAdapter implements the Context interface.
func (app *MainApplication) ModelAdapter() *model.Adapter {
	return app.modelAdapter
}

// OpenGl implements the Context interface.
func (app *MainApplication) OpenGl() opengl.OpenGl {
	return app.gl
}

// ForGraphics implements the Context interface.
func (app *MainApplication) ForGraphics() graphics.Context {
	return app
}

// RectangleRenderer implements the graphics.Context interface.
func (app *MainApplication) RectangleRenderer() *graphics.RectangleRenderer {
	return app.rectRenderer
}

// TextPainter implements the graphics.Context interface.
func (app *MainApplication) TextPainter() graphics.TextPainter {
	return app.defaultFontPainter
}

// Texturize implements the graphics.Context interface.
func (app *MainApplication) Texturize(bmp *graphics.Bitmap) *graphics.BitmapTexture {
	return graphics.NewBitmapTexture(app.gl, bmp.Width, bmp.Height, bmp.Pixels)
}

// UITextRenderer implements the graphics.Context interface.
func (app *MainApplication) UITextRenderer() *graphics.BitmapTextureRenderer {
	return app.uiTextRenderer
}

// ControlFactory implements the Context interface.
func (app *MainApplication) ControlFactory() controls.Factory {
	return app
}

// ForLabel implements the controls.Factory interface.
func (app *MainApplication) ForLabel() *controls.LabelBuilder {
	builder := controls.NewLabelBuilder(app.defaultFontPainter, app.Texturize, app.uiTextRenderer)
	builder.SetScale(2.0)
	return builder
}

// ForTextButton implements the controls.Factory interface.
func (app *MainApplication) ForTextButton() *controls.TextButtonBuilder {
	return controls.NewTextButtonBuilder(app.ForLabel(), app.rectRenderer)
}

// ForComboBox implements the controls.Factory interface.
func (app *MainApplication) ForComboBox() *controls.ComboBoxBuilder {
	return controls.NewComboBoxBuilder(app.ForLabel(), app.rectRenderer)
}

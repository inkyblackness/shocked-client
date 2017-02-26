package editor

import (
	"fmt"
	"os"
	"time"

	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/env/keys"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/opengl"
	"github.com/inkyblackness/shocked-client/ui"
	"github.com/inkyblackness/shocked-client/ui/events"
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

	mouseX, mouseY float32
	mouseButtons   uint32

	rootArea           *ui.Area
	defaultFontPainter graphics.TextPainter
	uiTextPalette      *graphics.PaletteTexture
	rectRenderer       *graphics.RectangleRenderer
	uiTextRenderer     *graphics.BitmapTextureRenderer
}

// NewMainApplication returns a new instance of MainApplication.
func NewMainApplication(store DataStore) *MainApplication {
	app := &MainApplication{
		projectionMatrix:   mgl.Ident4(),
		lastElapsedTick:    time.Now(),
		store:              store,
		defaultFontPainter: graphics.NewBitmapTextPainter(defaultFont)}

	return app
}

// Init implements the env.Application interface.
func (app *MainApplication) Init(glWindow env.OpenGlWindow) {
	app.setWindow(glWindow)
	app.setDebugOpenGl()
	app.initOpenGl()

	app.rectRenderer = graphics.NewRectangleRenderer(app.gl, &app.projectionMatrix)

	uiTextPalette := [][4]byte{
		{0x00, 0x00, 0x00, 0x00},
		{0xFF, 0x00, 0x00, 0xFF},
		{0x80, 0x00, 0x00, 0x40}}
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
	app.initInterface()

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

func (app *MainApplication) initInterface() {
	rootBuilder := ui.NewAreaBuilder()

	rootBuilder.SetRight(ui.NewAbsoluteAnchor(0.0))
	rootBuilder.SetBottom(ui.NewAbsoluteAnchor(0.0))
	app.rootArea = rootBuilder.Build()

	{
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

		centerPanelBuilder.OnRender(func(area *ui.Area) {
			app.rectRenderer.Fill(area.Left().Value(), area.Top().Value(), area.Right().Value(), area.Bottom().Value(),
				graphics.RGBA(0.25, 0, 0.25, 0.75))
		})
		centerPanelBuilder.Build()
	}

	{
		windowBuilder := ui.NewAreaBuilder()
		windowBuilder.SetParent(app.rootArea)

		windowHorizontalCenter := ui.NewOffsetAnchor(app.rootArea.Left(), 200.0)
		windowVerticalCenter := ui.NewRelativeAnchor(app.rootArea.Top(), app.rootArea.Bottom(), 0.5)

		windowBuilder.SetLeft(ui.NewOffsetAnchor(windowHorizontalCenter, -50.0))
		windowBuilder.SetRight(ui.NewOffsetAnchor(windowHorizontalCenter, 50.0))
		windowBuilder.SetTop(ui.NewOffsetAnchor(windowVerticalCenter, -50.0))
		windowBuilder.SetBottom(ui.NewOffsetAnchor(windowVerticalCenter, 50.0))

		lastGrabX, lastGrabY := float32(0.0), float32(0.0)

		windowBuilder.OnEvent(events.MouseButtonDownEventType, func(area *ui.Area, event events.Event) bool {
			buttonEvent := event.(*events.MouseButtonEvent)
			if buttonEvent.Buttons() == env.MousePrimary {
				area.RequestFocus()
				lastGrabX, lastGrabY = buttonEvent.Position()
			}
			return true
		})
		windowBuilder.OnEvent(events.MouseButtonUpEventType, func(area *ui.Area, event events.Event) bool {
			buttonEvent := event.(*events.MouseButtonEvent)
			if buttonEvent.AffectedButtons() == env.MousePrimary {
				area.ReleaseFocus()
			}
			return true
		})
		windowBuilder.OnEvent(events.MouseMoveEventType, func(area *ui.Area, event events.Event) bool {
			moveEvent := event.(*events.MouseMoveEvent)
			if area.HasFocus() {
				newX, newY := moveEvent.Position()
				windowHorizontalCenter.RequestValue(windowHorizontalCenter.Value() + (newX - lastGrabX))
				windowVerticalCenter.RequestValue(windowVerticalCenter.Value() + (newY - lastGrabY))
				lastGrabX, lastGrabY = newX, newY
			}
			return true
		})

		testTextBitmap := app.defaultFontPainter.Paint("Hello, World!\nSecond Line")
		textTexture := graphics.NewBitmapTexture(app.gl, testTextBitmap.Width, testTextBitmap.Height, testTextBitmap.Pixels)

		lineHeight := testTextBitmap.LineHeight()
		cursorLine := 1
		charOffset := testTextBitmap.CharOffset(cursorLine, 4)
		textScale := 2

		windowBuilder.OnRender(func(area *ui.Area) {
			app.rectRenderer.Fill(area.Left().Value(), area.Top().Value(), area.Right().Value(), area.Bottom().Value(),
				graphics.RGBA(1.0, 1.0, 1.0, 0.75))

			u, v := textTexture.UV()
			textWidth, textHeight := textTexture.Size()
			// graphics.RectByCoord(area.Left().Value(), area.Top().Value(), area.Left().Value()+textWidth*2, area.Top().Value()+textHeight*2)
			modelMatrix := mgl.Ident4().Mul4(mgl.Translate3D(area.Left().Value(), area.Top().Value(), 0.0)).
				Mul4(mgl.Scale3D(textWidth*float32(textScale), textHeight*float32(textScale), 1.0))
			app.uiTextRenderer.Render(&modelMatrix,
				textTexture,
				graphics.RectByCoord(0.0, 0.0, u, v))

			cursorTop := area.Top().Value() + float32(lineHeight*textScale*cursorLine)
			app.rectRenderer.Fill(area.Left().Value()+float32(charOffset*textScale), cursorTop,
				area.Left().Value()+float32(charOffset*textScale+2), cursorTop+float32(lineHeight*textScale),
				graphics.RGBA(0.0, 1.0, 0.0, 0.9))
		})

		windowArea := windowBuilder.Build()

		labelBuilder := controls.NewLabelBuilder(app.defaultFontPainter, app.texturize, app.uiTextRenderer)
		labelBuilder.SetParent(windowArea)
		labelBuilder.SetLeft(ui.NewOffsetAnchor(windowArea.Left(), 0))
		labelBuilder.SetTop(ui.NewOffsetAnchor(windowArea.Top(), 0))
		labelBuilder.SetRight(ui.NewOffsetAnchor(windowArea.Right(), 0))
		labelBuilder.SetBottom(ui.NewOffsetAnchor(windowArea.Bottom(), 0))
		label := labelBuilder.Build()
		label.SetText("I'm centered and a really long text that should be clipped")
	}
	{
		buttonLabelBuilder := controls.NewLabelBuilder(app.defaultFontPainter, app.texturize, app.uiTextRenderer)
		buttonLabelBuilder.SetScale(2.0)
		buttonBuilder := controls.NewTextButtonBuilder(buttonLabelBuilder, app.rectRenderer)
		buttonLeft := ui.NewOffsetAnchor(app.rootArea.Left(), 50)
		buttonTop := ui.NewOffsetAnchor(app.rootArea.Top(), 50)
		buttonBuilder.SetParent(app.rootArea)
		buttonBuilder.SetLeft(buttonLeft)
		buttonBuilder.SetTop(buttonTop)
		buttonBuilder.SetRight(ui.NewOffsetAnchor(buttonLeft, 100))
		buttonBuilder.SetBottom(ui.NewOffsetAnchor(buttonTop, 25))
		buttonBuilder.WithText("Click Me!")
		buttonBuilder.Build()
	}
}

func (app *MainApplication) texturize(bmp *graphics.Bitmap) *graphics.BitmapTexture {
	return graphics.NewBitmapTexture(app.gl, bmp.Width, bmp.Height, bmp.Pixels)
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
	app.rootArea.DispatchPositionalEvent(events.NewMouseMoveEvent(x, y, 0, 0))
}

func (app *MainApplication) onMouseButtonDown(mouseButton uint32, modifier keys.Modifier) {
	app.mouseButtons |= mouseButton
	app.rootArea.DispatchPositionalEvent(events.NewMouseButtonEvent(events.MouseButtonDownEventType,
		app.mouseX, app.mouseY, 0, app.mouseButtons, mouseButton))
}

func (app *MainApplication) onMouseButtonUp(mouseButton uint32, modifier keys.Modifier) {
	app.mouseButtons &= ^mouseButton
	app.rootArea.DispatchPositionalEvent(events.NewMouseButtonEvent(events.MouseButtonUpEventType,
		app.mouseX, app.mouseY, 0, app.mouseButtons, mouseButton))
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

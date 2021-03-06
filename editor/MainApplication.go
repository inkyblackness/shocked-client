package editor

import (
	"fmt"
	"os"
	"time"

	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/editor/cmd"
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

	commandStack cmd.Stack

	store        dataModel.DataStore
	modelAdapter *model.Adapter

	scale                float32
	invertedSliderScroll bool
	glWindow             env.OpenGlWindow
	gl                   opengl.OpenGl
	projectionMatrix     mgl.Mat4

	mouseX, mouseY      float32
	mouseButtons        uint32
	mouseButtonsDragged uint32
	keyModifier         keys.Modifier

	root               *rootArea
	rootArea           *ui.Area
	defaultFontPainter graphics.TextPainter
	uiTextPalette      *graphics.PaletteTexture
	rectRenderer       *graphics.RectangleRenderer
	uiTextRenderer     *graphics.BitmapTextureRenderer

	bitmaps              *graphics.BufferedTextureStore
	worldTextures        map[dataModel.TextureSize]*graphics.BufferedTextureStore
	gameObjectBitmaps    *graphics.BufferedTextureStore
	gameObjectIcons      *graphics.BufferedTextureStore
	worldPalette         *graphics.PaletteTexture
	worldTextureRenderer *graphics.BitmapTextureRenderer
}

// NewMainApplication returns a new instance of MainApplication.
func NewMainApplication(store dataModel.DataStore, scale float32, invertedSliderScroll bool) *MainApplication {
	app := &MainApplication{
		projectionMatrix:     mgl.Ident4(),
		lastElapsedTick:      time.Now(),
		store:                store,
		scale:                scale,
		invertedSliderScroll: invertedSliderScroll,
		modelAdapter:         model.NewAdapter(store),
		defaultFontPainter:   graphics.NewBitmapTextPainter(defaultFont),
		worldTextures:        make(map[dataModel.TextureSize]*graphics.BufferedTextureStore)}

	return app
}

// Init implements the env.Application interface.
func (app *MainApplication) Init(glWindow env.OpenGlWindow) {
	app.setWindow(glWindow)
	app.setDebugOpenGl()
	app.initOpenGl()

	app.initResources()
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
	glWindow.OnFileDropCallback(app.onFileDrop)
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

func (app *MainApplication) initResources() {
	for _, size := range dataModel.TextureSizes() {
		app.initWorldTextureBuffer(size)
	}
	app.initBitmaps()
	app.initGameObjectBitmapsBuffer()
	app.initWorldPalette()
}

func (app *MainApplication) initWorldTextureBuffer(size dataModel.TextureSize) {
	textureAdapter := app.modelAdapter.TextureAdapter()
	app.worldTextures[size] = app.createTextureStore(textureAdapter.WorldTextures(size), func(keyAsInt int) {
		textureAdapter.RequestWorldTextureBitmaps(keyAsInt)
	})
}

func (app *MainApplication) initBitmaps() {
	bitmapsAdapter := app.modelAdapter.BitmapsAdapter()
	app.bitmaps = app.createTextureStore(bitmapsAdapter.Bitmaps(), func(keyAsInt int) {
		bitmapsAdapter.RequestBitmap(dataModel.ResourceKeyFromInt(keyAsInt))
	})
}

func (app *MainApplication) initGameObjectBitmapsBuffer() {
	objectsAdapter := app.modelAdapter.ObjectsAdapter()
	app.gameObjectIcons = app.createTextureStore(objectsAdapter.Icons(), func(keyAsInt int) {
		objectsAdapter.RequestIcon(model.ObjectIDFromInt(keyAsInt))
	})
	app.gameObjectBitmaps = app.createTextureStore(objectsAdapter.Bitmaps(), func(keyAsInt int) {
		objectsAdapter.RequestBitmap(model.ObjectBitmapIDFromInt(keyAsInt))
	})
}

func (app *MainApplication) createTextureStore(bitmaps *model.Bitmaps, request func(int)) (buffer *graphics.BufferedTextureStore) {
	observedItems := make(map[int]bool)

	buffer = graphics.NewBufferedTextureStore(func(key graphics.TextureKey) {
		keyAsInt := key.ToInt()

		if !observedItems[keyAsInt] {
			bitmaps.OnBitmapChanged(keyAsInt, func() {
				raw := bitmaps.RawBitmap(keyAsInt)
				bmp := graphics.BitmapFromRaw(*raw)
				buffer.SetTexture(key, app.Texturize(&bmp))
			})
			observedItems[keyAsInt] = true
		}
		request(keyAsInt)
	})
	return
}

func (app *MainApplication) initWorldPalette() {
	gamePalette := app.modelAdapter.GamePalette()
	app.modelAdapter.OnGamePaletteChanged(func() {
		gamePalette = app.modelAdapter.GamePalette()
		app.worldPalette.Update()
	})
	app.worldPalette = graphics.NewPaletteTexture(app.gl, func(index int) (r byte, g byte, b byte, a byte) {
		color := &gamePalette[index]

		r = byte(color.Red)
		g = byte(color.Green)
		b = byte(color.Blue)
		if index > 0 {
			a = 0xFF
		}

		return
	})
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
	app.worldTextureRenderer = graphics.NewBitmapTextureRenderer(uiRenderContext, app.worldPalette)

	app.root, app.rootArea = newRootArea(app)
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
	app.mouseButtonsDragged |= app.mouseButtons
	app.rootArea.DispatchPositionalEvent(events.NewMouseMoveEvent(
		app.mouseX, app.mouseY, uint32(app.keyModifier), app.mouseButtons))
}

func (app *MainApplication) onMouseButtonDown(mouseButton uint32, modifier keys.Modifier) {
	app.mouseButtons |= mouseButton
	app.mouseButtonsDragged &= ^mouseButton
	app.rootArea.DispatchPositionalEvent(events.NewMouseButtonEvent(events.MouseButtonDownEventType,
		app.mouseX, app.mouseY, uint32(app.keyModifier), app.mouseButtons, mouseButton))
}

func (app *MainApplication) onMouseButtonUp(mouseButton uint32, modifier keys.Modifier) {
	app.mouseButtons &= ^mouseButton
	if (app.mouseButtonsDragged & mouseButton) == 0 {
		app.rootArea.DispatchPositionalEvent(events.NewMouseButtonEvent(events.MouseButtonClickedEventType,
			app.mouseX, app.mouseY, uint32(app.keyModifier), app.mouseButtons, mouseButton))
	}
	app.rootArea.DispatchPositionalEvent(events.NewMouseButtonEvent(events.MouseButtonUpEventType,
		app.mouseX, app.mouseY, uint32(app.keyModifier), app.mouseButtons, mouseButton))
}

func (app *MainApplication) onMouseScroll(dx float32, dy float32) {
	app.rootArea.DispatchPositionalEvent(events.NewMouseScrollEvent(
		app.mouseX, app.mouseY, uint32(app.keyModifier), app.mouseButtons, dx, dy))
}

func (app *MainApplication) onKey(key keys.Key, modifier keys.Modifier) {
	app.keyModifier = modifier
	if key == keys.KeySave {
		app.modelAdapter.SaveProject()
	} else if key == keys.KeyCopy {
		app.rootArea.DispatchPositionalEvent(events.NewClipboardEvent(events.ClipboardCopyEventType,
			app.mouseX, app.mouseY, app.glWindow.Clipboard()))
	} else if key == keys.KeyPaste {
		app.rootArea.DispatchPositionalEvent(events.NewClipboardEvent(events.ClipboardPasteEventType,
			app.mouseX, app.mouseY, app.glWindow.Clipboard()))
	} else if key == keys.KeyUndo {
		app.undo()
	} else if key == keys.KeyRedo {
		app.redo()
	} else if (key >= keys.KeyF1) && (key <= keys.KeyF9) {
		modeIndex := key - keys.KeyF1
		app.root.RequestActiveMode(app.root.ModeNames()[modeIndex])
	}
}

func (app *MainApplication) onModifier(modifier keys.Modifier) {
	app.keyModifier = modifier
}

func (app *MainApplication) onChar(char rune) {
}

func (app *MainApplication) onFileDrop(filePaths []string) {
	app.sendFileDropEvent(filePaths)
}

func (app *MainApplication) sendFileDropEvent(filePaths []string) {
	app.rootArea.DispatchPositionalEvent(events.NewFileDropEvent(
		app.mouseX, app.mouseY, filePaths))
}

func (app *MainApplication) undo() {
	err := app.commandStack.Undo()
	if err != nil {
		app.modelAdapter.SetMessage("Failed to undo command.")
	}
}

func (app *MainApplication) redo() {
	err := app.commandStack.Redo()
	if err != nil {
		app.modelAdapter.SetMessage("Failed to redo command.")
	}
}

// Perform tries to execute the given command and puts it on the command stack.
func (app *MainApplication) Perform(command cmd.Command) {
	err := app.commandStack.Perform(command)
	if err != nil {
		app.modelAdapter.SetMessage("Failed to perform command.")
	}
}

// ModelAdapter implements the Context interface.
func (app *MainApplication) ModelAdapter() *model.Adapter {
	return app.modelAdapter
}

// NewRenderContext implements the Context interface.
func (app *MainApplication) NewRenderContext(viewMatrix *mgl.Mat4) *graphics.RenderContext {
	return graphics.NewBasicRenderContext(app.gl, &app.projectionMatrix, viewMatrix)
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

// NewPaletteTexture implements the graphics.Context interface.
func (app *MainApplication) NewPaletteTexture(colorProvider graphics.ColorProvider) *graphics.PaletteTexture {
	return graphics.NewPaletteTexture(app.gl, colorProvider)
}

// WorldTextureStore implements the graphics.Context interface.
func (app *MainApplication) WorldTextureStore(size dataModel.TextureSize) *graphics.BufferedTextureStore {
	return app.worldTextures[size]
}

// GameObjectBitmapsStore implements the graphics.Context interface.
func (app *MainApplication) GameObjectBitmapsStore() *graphics.BufferedTextureStore {
	return app.gameObjectBitmaps
}

// GameObjectIconsStore implements the graphics.Context interface.
func (app *MainApplication) GameObjectIconsStore() *graphics.BufferedTextureStore {
	return app.gameObjectIcons
}

// BitmapsStore implements the graphics.Context interface.
func (app *MainApplication) BitmapsStore() *graphics.BufferedTextureStore {
	return app.bitmaps
}

// ControlFactory implements the Context interface.
func (app *MainApplication) ControlFactory() controls.Factory {
	return app
}

// Scale implements the controls.Factory interface.
func (app *MainApplication) Scale() float32 {
	return app.scale
}

// ForLabel implements the controls.Factory interface.
func (app *MainApplication) ForLabel() *controls.LabelBuilder {
	builder := controls.NewLabelBuilder(app.defaultFontPainter, app.Texturize, app.uiTextRenderer)
	builder.SetScale(2.0 * app.Scale())
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

// ForTextureSelector implements the controls.Factory interface.
func (app *MainApplication) ForTextureSelector() *controls.TextureSelectorBuilder {
	return controls.NewTextureSelectorBuilder(app.rectRenderer, app.worldTextureRenderer).WithInvertedScroll(app.invertedSliderScroll)
}

// ForSlider implements the controls.Factory interface.
func (app *MainApplication) ForSlider() *controls.SliderBuilder {
	return controls.NewSliderBuilder(app.ForLabel(), app.rectRenderer).WithInvertedScroll(app.invertedSliderScroll)
}

// ForImageDisplay implements the controls.Factory interface.
func (app *MainApplication) ForImageDisplay() *controls.ImageDisplayBuilder {
	return controls.NewImageDisplayBuilder(app.worldTextureRenderer)
}

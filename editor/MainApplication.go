package editor

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/editor/camera"
	"github.com/inkyblackness/shocked-client/editor/display"
	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/opengl"
	"github.com/inkyblackness/shocked-client/viewmodel"
	"github.com/inkyblackness/shocked-model"
)

// MainApplication represents the core intelligence of the editor.
type MainApplication struct {
	lastElapsedTick time.Time
	elapsedMSec     int64

	store DataStore

	viewModel *ViewModel

	glWindow env.OpenGlWindow
	gl       opengl.OpenGl

	mouseX, mouseY   float32
	mouseMoveCapture func()

	view *camera.LimitedCamera

	paletteTexture           *graphics.PaletteTexture
	gridRenderable           *display.GridRenderable
	tileTextureMapRenderable *display.TileTextureMapRenderable
}

// NewMainApplication returns a new instance of MainApplication.
func NewMainApplication(store DataStore) *MainApplication {
	camLimit := (TilesPerMapSide - 1) * TileBaseLength
	app := &MainApplication{
		lastElapsedTick:  time.Now(),
		store:            store,
		viewModel:        NewViewModel(),
		mouseMoveCapture: func() {},
		view:             camera.NewLimited(ZoomLevelMin, ZoomLevelMax, 0, camLimit)}

	app.viewModel.OnSelectedProjectChanged(app.onSelectedProjectChanged)
	app.viewModel.OnSelectedLevelChanged(app.onSelectedLevelChanged)
	store.Projects(func(projectIDs []string) {
		app.viewModel.SetProjects(projectIDs)
	}, app.simpleStoreFailure("Projects"))

	return app
}

// ViewModel implements the env.Application interface.
func (app *MainApplication) ViewModel() viewmodel.Node {
	return app.viewModel.Root()
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

	//app.gl = app.glWindow.OpenGl()
	app.gl = builder.Build()

	app.gl.Enable(opengl.BLEND)
	app.gl.BlendFunc(opengl.SRC_ALPHA, opengl.ONE_MINUS_SRC_ALPHA)
	app.gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	app.gridRenderable = display.NewGridRenderable(app.gl)
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

func (app *MainApplication) render() {
	gl := app.gl
	width, height := app.glWindow.Size()

	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(opengl.COLOR_BUFFER_BIT | opengl.DEPTH_BUFFER_BIT)

	app.updateElapsedNano()
	if app.paletteTexture != nil {
		app.paletteTexture.Update()
	}
	context := display.NewBasicRenderContext(width, height, app.view.ViewMatrix())

	app.gridRenderable.Render(context)
	if app.tileTextureMapRenderable != nil {
		app.tileTextureMapRenderable.Render(context)
	}
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

func (app *MainApplication) animatedPaletteIndex(index int) int {
	newIndex := index
	loopIndex := func(from int, count int, stepTimeMSec int64) {
		if newIndex >= from && newIndex < (from+count) {
			step := app.elapsedMSec / stepTimeMSec
			newIndex = from + int(int64(newIndex-from)+step)%count
		}
	}
	loopIndex(0x03, 5, 1200)
	loopIndex(0x0B, 5, 700)
	loopIndex(0x10, 5, 360)
	loopIndex(0x15, 3, 1800)
	loopIndex(0x18, 3, 1430)
	loopIndex(0x1B, 5, 1080)

	return newIndex
}

func (app *MainApplication) onSelectedProjectChanged(projectID string) {
	app.viewModel.SetLevels(nil)

	if app.tileTextureMapRenderable != nil {
		app.tileTextureMapRenderable.Dispose()
		app.tileTextureMapRenderable = nil
	}
	if app.paletteTexture != nil {
		app.paletteTexture.Dispose()
		app.paletteTexture = nil
	}
	if projectID != "" {

		app.store.Palette(projectID, "game", func(colors [256]model.Color) {
			colorProvider := func(index int) (byte, byte, byte, byte) {
				entry := &colors[app.animatedPaletteIndex(index)]
				return byte(entry.Red), byte(entry.Green), byte(entry.Blue), 255
			}
			app.paletteTexture = graphics.NewPaletteTexture(app.gl, colorProvider)
			app.tileTextureMapRenderable = display.NewTileTextureMapRenderable(app.gl, app.paletteTexture)
		}, app.simpleStoreFailure("Palette"))

		app.store.Levels(projectID, "archive", func(levels []model.Level) {
			levelIDs := make([]string, len(levels))
			for index, level := range levels {
				levelIDs[index] = level.ID
			}
			app.viewModel.SetLevels(levelIDs)
		}, app.simpleStoreFailure("Levels"))
	}
}

func (app *MainApplication) onSelectedLevelChanged(levelIDString string) {
	projectID := app.viewModel.SelectedProject()
	levelID, levelIDError := strconv.ParseInt(levelIDString, 10, 16)

	if app.tileTextureMapRenderable != nil {
		app.tileTextureMapRenderable.Clear()
	}
	if projectID != "" && levelIDError == nil {
		var levelTextureIDs []int
		bitmapTextures := make(map[int]graphics.Texture)
		var tiles *model.Tiles

		createMap := func() {
			if tiles != nil &&
				len(levelTextureIDs) > 0 && len(bitmapTextures) == len(levelTextureIDs) &&
				app.tileTextureMapRenderable != nil {

				for y := 0; y < len(tiles.Table); y++ {
					row := tiles.Table[y]
					for x := 0; x < len(row); x++ {
						tileData := &row[x]

						if *tileData.Properties.Type != model.Solid {
							textureID := *tileData.Properties.RealWorld.FloorTexture

							app.tileTextureMapRenderable.SetTileTexture(x, 63-y, bitmapTextures[textureID])
						}
					}
				}
			}
		}

		app.store.Tiles(projectID, "archive", int(levelID), func(data model.Tiles) {
			tiles = &data

			createMap()
		}, app.simpleStoreFailure("Tiles"))

		app.store.LevelTextures(projectID, "archive", int(levelID), func(textureIDs []int) {
			fmt.Fprintf(os.Stderr, "loaded textureIDs, %v to load\n", len(textureIDs))
			levelTextureIDs = textureIDs
			for index, id := range textureIDs {
				localIndex := index
				app.store.TextureBitmap(projectID, id, "large", func(bmp *model.RawBitmap) {
					pixelData, _ := base64.StdEncoding.DecodeString(bmp.Pixel)
					bitmapTextures[localIndex] = graphics.NewBitmapTexture(app.gl, bmp.Width, bmp.Height, pixelData)

					createMap()
				}, app.simpleStoreFailure("TextureBitmap"))
			}
		}, app.simpleStoreFailure("LevelTextures"))
	}
}

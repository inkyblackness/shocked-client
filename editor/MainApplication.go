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
	editormodel "github.com/inkyblackness/shocked-client/editor/model"
	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/opengl"
	"github.com/inkyblackness/shocked-client/util"
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
	mouseDragged     bool
	mouseMoveCapture func()

	view *camera.LimitedCamera

	paletteTexture *graphics.PaletteTexture
	levelTextures  []int
	textureStore   *editormodel.BufferedTextureStore
	tileMap        *editormodel.TileMap

	gridRenderable           *display.GridRenderable
	tileTextureMapRenderable *display.TileTextureMapRenderable
	tileGridMapRenderable    *display.TileGridMapRenderable
	tileSelectionRenderable  *display.TileSelectionRenderable
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

	app.textureStore = editormodel.NewBufferedTextureStore(app.loadTexture)
	app.tileMap = editormodel.NewTileMap(TilesPerMapSide, TilesPerMapSide)

	store.Projects(func(projectIDs []string) {
		app.viewModel.SetProjects(projectIDs)
		if (len(projectIDs) == 1) && (projectIDs[0] == "(inplace)") {
			app.viewModel.SelectProject("(inplace)")
			app.viewModel.SelectMapSection()
		}
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
	app.tileSelectionRenderable = display.NewTileSelectionRenderable(app.gl, func(callback display.TileSelectionCallback) {
		app.tileMap.ForEachSelected(func(coord editormodel.TileCoordinate, tile *editormodel.Tile) {
			callback(coord)
		})
	})
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
	app.tileSelectionRenderable.Render(context)
	if app.tileGridMapRenderable != nil {
		app.tileGridMapRenderable.Render(context)
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

	app.mouseMoveCapture()

	worldMouseX, worldMouseY := app.unprojectPixel(app.mouseX, app.mouseY)
	tileX, subX := int(worldMouseX/TileBaseLength), (int(worldMouseX/TileBaseLength*256.0))%256
	tileY, subY := int(TilesPerMapSide)-1-int(worldMouseY/TileBaseLength), 255-((int(worldMouseY/TileBaseLength*256.0))%256)
	app.viewModel.SetPointerAt(tileX, tileY, subX, subY)
}

func (app *MainApplication) onMouseButtonDown(mouseButton uint32, modifierMask uint32) {
	if (mouseButton & env.MousePrimary) == env.MousePrimary {
		lastMouseX, lastMouseY := app.mouseX, app.mouseY

		app.mouseDragged = false
		app.mouseMoveCapture = func() {
			lastWorldMouseX, lastWorldMouseY := app.unprojectPixel(lastMouseX, lastMouseY)
			worldMouseX, worldMouseY := app.unprojectPixel(app.mouseX, app.mouseY)

			app.mouseDragged = true
			app.view.MoveBy(worldMouseX-lastWorldMouseX, worldMouseY-lastWorldMouseY)
			lastMouseX, lastMouseY = app.mouseX, app.mouseY
		}
	}
}

func (app *MainApplication) onMouseButtonUp(mouseButton uint32, modifierMask uint32) {
	if (mouseButton & env.MousePrimary) == env.MousePrimary {
		app.mouseMoveCapture = func() {}
		if !app.mouseDragged {
			app.onMouseClick(modifierMask)
		}
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

func (app *MainApplication) onMouseClick(modifierMask uint32) {
	worldMouseX, worldMouseY := app.unprojectPixel(app.mouseX, app.mouseY)
	tileX, _ := int(worldMouseX/TileBaseLength), (int(worldMouseX/TileBaseLength*256.0))%256
	tileY, _ := int(TilesPerMapSide)-1-int(worldMouseY/TileBaseLength), 255-((int(worldMouseY/TileBaseLength*256.0))%256)

	tileCoord := editormodel.TileCoordinateOf(tileX, tileY)
	if (modifierMask & env.ModControl) != 0 {
		app.tileMap.SetSelected(tileCoord, !app.tileMap.IsSelected(tileCoord))
	} else {
		app.tileMap.ClearSelection()
		app.tileMap.SetSelected(tileCoord, true)
	}
	app.onTileSelectionChanged()
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
	if app.tileGridMapRenderable != nil {
		app.tileGridMapRenderable.Dispose()
		app.tileGridMapRenderable = nil
	}
	if app.paletteTexture != nil {
		app.paletteTexture.Dispose()
		app.paletteTexture = nil
	}
	app.textureStore.Reset()
	if projectID != "" {

		app.store.Palette(projectID, "game", func(colors [256]model.Color) {
			colorProvider := func(index int) (byte, byte, byte, byte) {
				entry := &colors[app.animatedPaletteIndex(index)]
				return byte(entry.Red), byte(entry.Green), byte(entry.Blue), 255
			}
			app.paletteTexture = graphics.NewPaletteTexture(app.gl, colorProvider)
			app.tileTextureMapRenderable = display.NewTileTextureMapRenderable(app.gl, app.paletteTexture, app.levelTexture)
			app.tileGridMapRenderable = display.NewTileGridMapRenderable(app.gl)
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
	if app.tileGridMapRenderable != nil {
		app.tileGridMapRenderable.Clear()
	}
	if projectID != "" && levelIDError == nil {
		app.store.Tiles(projectID, "archive", int(levelID), func(data model.Tiles) {
			for y, row := range data.Table {
				for x := 0; x < len(row); x++ {
					coord := editormodel.TileCoordinateOf(x, y)
					properties := &row[x].Properties
					app.tileTextureMapRenderable.SetTile(x, 63-y, properties)
					app.tileGridMapRenderable.SetTile(x, 63-y, properties)
					app.tileMap.Tile(coord).SetProperties(properties)
				}
			}
		}, app.simpleStoreFailure("Tiles"))

		app.store.LevelTextures(projectID, "archive", int(levelID), func(textureIDs []int) {
			app.levelTextures = textureIDs
		}, app.simpleStoreFailure("LevelTextures"))
	}
}

func (app *MainApplication) loadTexture(id int) {
	projectID := app.viewModel.SelectedProject()

	app.store.TextureBitmap(projectID, id, "large", func(bmp *model.RawBitmap) {
		pixelData, _ := base64.StdEncoding.DecodeString(bmp.Pixel)
		app.textureStore.SetTexture(id, graphics.NewBitmapTexture(app.gl, bmp.Width, bmp.Height, pixelData))
	}, app.simpleStoreFailure("TextureBitmap"))
}

func (app *MainApplication) levelTexture(index int) (texture graphics.Texture) {
	if index >= 0 && index < len(app.levelTextures) {
		texture = app.textureStore.Texture(app.levelTextures[index])
	}

	return
}

func (app *MainApplication) onTileSelectionChanged() {
	tileType := util.NewValueUnifier("")

	app.tileMap.ForEachSelected(func(coord editormodel.TileCoordinate, tile *editormodel.Tile) {
		tileType.Add(string(*tile.Properties().Type))
	})

	app.viewModel.Tiles().TileType().Selected().Set(tileType.Value().(string))
}

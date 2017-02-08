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

	"github.com/inkyblackness/res"
	"github.com/inkyblackness/res/data/interpreters"
	"github.com/inkyblackness/res/data/levelobj"
)

// MainApplication represents the core intelligence of the editor.
type MainApplication struct {
	lastElapsedTick time.Time
	elapsedMSec     int64

	store DataStore

	viewModel         *ViewModel
	viewModelUpdating bool

	glWindow env.OpenGlWindow
	gl       opengl.OpenGl

	mouseX, mouseY   float32
	mouseDragged     bool
	mouseMoveCapture func()

	view *camera.LimitedCamera

	gameObjectIcons         map[editormodel.ObjectID]*graphics.BitmapTexture
	gameObjectIconRetriever map[editormodel.ObjectID]graphics.BitmapRetriever

	levels             []model.Level
	activeLevelID      int
	paletteTexture     *graphics.PaletteTexture
	levelTextures      []int
	textureData        []model.Texture
	gameTextureStore   *editormodel.BufferedTextureStore
	tileMap            *editormodel.TileMap
	levelObjects       map[int]*editormodel.LevelObject
	levelObjectIndices []int

	defaultFont graphics.TextRenderer
	defaultIcon *graphics.BitmapTexture

	gridRenderable           *display.GridRenderable
	tileTextureMapRenderable *display.TileTextureMapRenderable
	tileGridMapRenderable    *display.TileGridMapRenderable
	tileSelectionRenderable  *display.TileSelectionRenderable

	simpleBitmapRenderable *display.SimpleBitmapRenderable
	highlighter            *display.BasicHighlighter
}

// NewMainApplication returns a new instance of MainApplication.
func NewMainApplication(store DataStore) *MainApplication {
	camLimit := (TilesPerMapSide - 1) * TileBaseLength
	app := &MainApplication{
		lastElapsedTick:         time.Now(),
		store:                   store,
		viewModel:               NewViewModel(),
		mouseMoveCapture:        func() {},
		view:                    camera.NewLimited(ZoomLevelMin, ZoomLevelMax, 0, camLimit),
		gameObjectIcons:         make(map[editormodel.ObjectID]*graphics.BitmapTexture),
		gameObjectIconRetriever: make(map[editormodel.ObjectID]graphics.BitmapRetriever),
		levelObjects:            make(map[int]*editormodel.LevelObject)}

	app.viewModel.OnSelectedProjectChanged(app.onSelectedProjectChanged)
	app.viewModel.CreateProject().Subscribe(app.onCreateProject)
	app.viewModel.OnSelectedLevelChanged(app.onSelectedLevelChanged)
	app.viewModel.Tiles().TileType().Selected().Subscribe(app.onTileTypeChanged)
	app.viewModel.Tiles().FloorHeight().Selected().Subscribe(app.onTileFloorHeightChanged)
	app.viewModel.Tiles().CeilingHeight().Selected().Subscribe(app.onTileCeilingHeightChanged)
	app.viewModel.Tiles().SlopeHeight().Selected().Subscribe(app.onTileSlopeHeightChanged)
	app.viewModel.Tiles().SlopeControl().Selected().Subscribe(app.onTileSlopeControlChanged)
	app.viewModel.Tiles().FloorTexture().Selected().Subscribe(app.tileIntRealWorldValueChangeCallback(func(properties *model.RealWorldTileProperties) **int {
		return &properties.FloorTexture
	}, false))
	app.viewModel.Tiles().CeilingTexture().Selected().Subscribe(app.tileIntRealWorldValueChangeCallback(func(properties *model.RealWorldTileProperties) **int {
		return &properties.CeilingTexture
	}, false))
	app.viewModel.Tiles().WallTexture().Selected().Subscribe(app.tileIntRealWorldValueChangeCallback(func(properties *model.RealWorldTileProperties) **int {
		return &properties.WallTexture
	}, false))
	app.viewModel.Tiles().FloorTextureRotations().Selected().Subscribe(app.tileIntRealWorldValueChangeCallback(func(properties *model.RealWorldTileProperties) **int {
		return &properties.FloorTextureRotations
	}, false))
	app.viewModel.Tiles().CeilingTextureRotations().Selected().Subscribe(app.tileIntRealWorldValueChangeCallback(func(properties *model.RealWorldTileProperties) **int {
		return &properties.CeilingTextureRotations
	}, false))
	app.viewModel.Tiles().WallTextureOffset().Selected().Subscribe(app.onTileWallTextureOffsetChanged)
	app.viewModel.Tiles().UseAdjacentWallTexture().Selected().Subscribe(app.onTileUseAdjacentWallTextureChanged)

	app.viewModel.LevelTextureIndex().Selected().Subscribe(app.onLevelTextureIndexChanged)
	app.viewModel.LevelTextureID().Selected().Subscribe(app.onLevelTextureIDChanged)

	app.viewModel.LevelObjects().SelectedObject().Selected().Subscribe(app.onSelectedObjectIndexChanged)

	app.activeLevelID = -1
	app.gameTextureStore = editormodel.NewBufferedTextureStore(app.loadGameTexture)
	app.tileMap = editormodel.NewTileMap(TilesPerMapSide, TilesPerMapSide)

	app.queryProjectsAndSelect("(inplace)")

	return app
}

func (app *MainApplication) queryProjectsAndSelect(projectID string) {
	app.store.Projects(func(projectIDs []string) {
		app.viewModel.SetProjects(projectIDs)

		found := false
		for _, id := range projectIDs {
			if id == projectID {
				found = true
			}
		}
		if found {
			app.viewModel.SelectProject(projectID)
			app.viewModel.SelectMapSection()
		}
	}, app.simpleStoreFailure("Projects"))
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

	app.highlighter = display.NewBasicHighlighter(app.gl, [4]float32{1.0, 1.0, 1.0, 0.4})
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

	areaList := make([]display.Area, 0, len(app.levelObjects))
	iconList := make([]display.PlacedIcon, 0, len(app.levelObjects))
	for _, objectIndex := range app.levelObjectIndices {
		obj := app.levelObjects[objectIndex]
		areaList = append(areaList, obj)
		iconList = append(iconList, obj)
	}
	app.highlighter.Render(context, areaList)
	if app.simpleBitmapRenderable != nil {
		app.simpleBitmapRenderable.Render(context, iconList)
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

func (app *MainApplication) onCreateProject() {
	projectID := app.viewModel.NewProjectID().Get()

	if projectID != "" {
		app.store.NewProject(projectID, func() {
			app.viewModel.NewProjectID().Set("")
			app.queryProjectsAndSelect(projectID)
		}, app.simpleStoreFailure("NewProject"))
	}
}

func (app *MainApplication) onSelectedProjectChanged(projectID string) {
	app.updateViewModel(func() {
		app.viewModel.SetLevels(nil)
		app.viewModel.SetTextureCount(0)
	})

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
	app.gameObjectIconRetriever = make(map[editormodel.ObjectID]graphics.BitmapRetriever)
	for _, texture := range app.gameObjectIcons {
		texture.Dispose()
	}
	app.gameObjectIcons = make(map[editormodel.ObjectID]*graphics.BitmapTexture)
	app.textureData = nil
	app.gameTextureStore.Reset()
	app.levels = nil

	if projectID != "" {
		app.store.Font(projectID, 0x025B, func(font *model.Font) {
			app.defaultFont = graphics.NewBitmapTextRenderer(*font)
			bmp := app.defaultFont.Render("?")
			app.defaultIcon = graphics.NewBitmapTexture(app.gl, bmp.Width, bmp.Height, bmp.Pixels)
		}, app.simpleStoreFailure("Font"))

		app.store.Palette(projectID, "game", func(colors [256]model.Color) {
			colorProvider := func(index int) (byte, byte, byte, byte) {
				entry := &colors[app.animatedPaletteIndex(index)]
				return byte(entry.Red), byte(entry.Green), byte(entry.Blue), 255
			}
			app.paletteTexture = graphics.NewPaletteTexture(app.gl, colorProvider)
			app.tileGridMapRenderable = display.NewTileGridMapRenderable(app.gl)

			app.simpleBitmapRenderable = display.NewSimpleBitmapRenderable(app.gl, app.paletteTexture)
		}, app.simpleStoreFailure("Palette"))

		app.store.Textures(projectID, func(textures []model.Texture) {
			app.textureData = textures
			app.updateViewModel(func() {
				app.viewModel.SetTextureCount(len(textures))
			})
		}, app.simpleStoreFailure("Textures"))

		app.store.Levels(projectID, "archive", func(levels []model.Level) {
			levelIDs := make([]string, len(levels))
			for index, level := range levels {
				levelIDs[index] = level.ID
			}
			app.levels = levels
			app.updateViewModel(func() {
				app.viewModel.SetLevels(levelIDs)
			})
		}, app.simpleStoreFailure("Levels"))
	}
}

func (app *MainApplication) onSelectedLevelChanged(levelIDString string) {
	projectID := app.viewModel.SelectedProject()
	levelID, levelIDError := strconv.ParseInt(levelIDString, 10, 16)

	if app.tileTextureMapRenderable != nil {
		app.tileTextureMapRenderable.Clear()
		app.tileTextureMapRenderable.Dispose()
		app.tileTextureMapRenderable = nil
	}
	if app.tileGridMapRenderable != nil {
		app.tileGridMapRenderable.Clear()
	}
	app.tileMap.Clear()
	app.onTileSelectionChanged()
	app.activeLevelID = -1
	app.updateViewModel(func() {
		app.viewModel.SetLevelTextures(nil)
	})

	if projectID != "" && levelIDError == nil {
		app.activeLevelID = int(levelID)

		if app.isActiveLevelRealWorld() {
			app.tileTextureMapRenderable = display.NewTileTextureMapRenderable(app.gl, app.paletteTexture, app.levelTexture)
		}

		app.store.Tiles(projectID, "archive", app.activeLevelID, func(data model.Tiles) {
			for y, row := range data.Table {
				for x := 0; x < len(row); x++ {
					coord := editormodel.TileCoordinateOf(x, y)
					properties := &row[x].Properties
					app.onTilePropertiesUpdated(coord, properties)
				}
			}
		}, app.simpleStoreFailure("Tiles"))

		app.store.LevelTextures(projectID, "archive", app.activeLevelID,
			app.onStoreLevelTexturesChanged, app.simpleStoreFailure("LevelTextures"))

		app.store.LevelObjects(projectID, "archive", app.activeLevelID,
			app.onStoreLevelObjectsChanged, app.simpleStoreFailure("LevelObjects"))
	}

	app.updateViewModel(func() {
		app.viewModel.SetLevelIsRealWorld(app.isActiveLevelRealWorld())
	})
}

func (app *MainApplication) activeLevel() (level *model.Level) {
	activeLevelIDString := fmt.Sprintf("%d", app.activeLevelID)

	for index := 0; index < len(app.levels); index++ {
		if app.levels[index].ID == activeLevelIDString {
			level = &app.levels[index]
		}
	}

	return
}

func (app *MainApplication) isActiveLevelRealWorld() (realWorld bool) {
	activeLevel := app.activeLevel()

	if activeLevel != nil {
		realWorld = !activeLevel.Properties.CyberspaceFlag
	}

	return
}

func (app *MainApplication) onStoreLevelTexturesChanged(textureIDs []int) {
	app.levelTextures = textureIDs
	app.updateViewModel(func() {
		app.viewModel.SetLevelTextures(app.levelTextures)
	})
}

func (app *MainApplication) loadGameTexture(id editormodel.TextureKey) {
	projectID := app.viewModel.SelectedProject()
	gameTextureKey := id.(editormodel.GameTextureKey)

	app.store.TextureBitmap(projectID, gameTextureKey.ID(), "large", func(bmp *model.RawBitmap) {
		pixelData, _ := base64.StdEncoding.DecodeString(bmp.Pixels)
		app.gameTextureStore.SetTexture(id, graphics.NewBitmapTexture(app.gl, bmp.Width, bmp.Height, pixelData))
	}, app.simpleStoreFailure("TextureBitmap"))
}

func (app *MainApplication) levelTexture(index int) (texture graphics.Texture) {
	if index >= 0 && index < len(app.levelTextures) {
		texture = app.gameTextureStore.Texture(editormodel.GameTextureKeyFor(app.levelTextures[index]))
	}

	return
}

func (app *MainApplication) onTilePropertiesUpdated(coord editormodel.TileCoordinate, properties *model.TileProperties) {
	x, y := coord.XY()

	if app.tileTextureMapRenderable != nil {
		app.tileTextureMapRenderable.SetTile(x, 63-y, properties)
	}
	app.tileGridMapRenderable.SetTile(x, 63-y, properties)
	app.tileMap.Tile(coord).SetProperties(properties)
}

func (app *MainApplication) updateViewModel(updater func()) {
	app.viewModelUpdating = true
	defer func() {
		app.viewModelUpdating = false
	}()

	updater()
}

func (app *MainApplication) requestSelectedTilesChange(modifier func(*model.TileProperties), updateNeighbours bool) {
	if !app.viewModelUpdating {
		projectID := app.viewModel.SelectedProject()
		archiveID := "archive"
		levelID := app.activeLevelID
		neighbours := make(map[editormodel.TileCoordinate]int)
		writesPending := 0

		onWriteCompleted := func() {
			writesPending--
			if writesPending == 0 {
				for coord := range neighbours {
					localCoord := coord
					x, y := localCoord.XY()
					app.store.Tile(projectID, archiveID, levelID, x, y, func(properties model.TileProperties) {
						app.onTilePropertiesUpdated(localCoord, &properties)
					}, app.simpleStoreFailure("GetTile"))
				}
			}
		}

		app.tileMap.ForEachSelected(func(coord editormodel.TileCoordinate, tile *editormodel.Tile) {
			var properties model.TileProperties

			modifier(&properties)

			writesPending++
			x, y := coord.XY()
			if updateNeighbours {
				if x > 0 {
					neighbours[editormodel.TileCoordinateOf(x-1, y)]++
				}
				if (x + 1) < TilesPerMapSide {
					neighbours[editormodel.TileCoordinateOf(x+1, y)]++
				}
				if y > 0 {
					neighbours[editormodel.TileCoordinateOf(x, y-1)]++
				}
				if (y + 1) < TilesPerMapSide {
					neighbours[editormodel.TileCoordinateOf(x, y+1)]++
				}
			}
			app.store.SetTile(projectID, archiveID, levelID, x, y, properties, func(newProperties model.TileProperties) {
				app.onTilePropertiesUpdated(coord, &newProperties)
				onWriteCompleted()
			}, app.simpleStoreFailure("SetTile"))
		})
	}
}

func (app *MainApplication) onTileSelectionChanged() {
	tileType := util.NewValueUnifier("")
	floorHeight := util.NewValueUnifier("")
	ceilingHeight := util.NewValueUnifier("")
	slopeHeight := util.NewValueUnifier("")
	slopeControl := util.NewValueUnifier("")
	floorTexture := util.NewValueUnifier("")
	ceilingTexture := util.NewValueUnifier("")
	wallTexture := util.NewValueUnifier("")
	floorTextureRotations := util.NewValueUnifier("")
	ceilingTextureRotations := util.NewValueUnifier("")
	useAdjacentWallTexture := util.NewValueUnifier("")
	wallTextureOffset := util.NewValueUnifier("")

	app.tileMap.ForEachSelected(func(coord editormodel.TileCoordinate, tile *editormodel.Tile) {
		tileType.Add(string(*tile.Properties().Type))
		floorHeight.Add(fmt.Sprintf("%d", *tile.Properties().FloorHeight))
		ceilingHeight.Add(fmt.Sprintf("%d", 32-*tile.Properties().CeilingHeight))
		slopeHeight.Add(fmt.Sprintf("%d", *tile.Properties().SlopeHeight))
		slopeControl.Add(string(*tile.Properties().SlopeControl))
		if tile.Properties().RealWorld != nil {
			realWorld := tile.Properties().RealWorld
			floorTexture.Add(fmt.Sprintf("%d", *realWorld.FloorTexture))
			ceilingTexture.Add(fmt.Sprintf("%d", *realWorld.CeilingTexture))
			wallTexture.Add(fmt.Sprintf("%d", *realWorld.WallTexture))
			floorTextureRotations.Add(fmt.Sprintf("%d", *realWorld.FloorTextureRotations))
			ceilingTextureRotations.Add(fmt.Sprintf("%d", *realWorld.CeilingTextureRotations))
			if *realWorld.UseAdjacentWallTexture {
				useAdjacentWallTexture.Add("yes")
			} else {
				useAdjacentWallTexture.Add("no")
			}
			wallTextureOffset.Add(fmt.Sprintf("%d", *realWorld.WallTextureOffset))
		}
	})

	app.updateViewModel(func() {
		app.viewModel.Tiles().TileType().Selected().Set(tileType.Value().(string))
		app.viewModel.Tiles().FloorHeight().Selected().Set(floorHeight.Value().(string))
		app.viewModel.Tiles().CeilingHeight().Selected().Set(ceilingHeight.Value().(string))
		app.viewModel.Tiles().SlopeHeight().Selected().Set(slopeHeight.Value().(string))
		app.viewModel.Tiles().SlopeControl().Selected().Set(slopeControl.Value().(string))
		app.viewModel.Tiles().FloorTexture().Selected().Set(floorTexture.Value().(string))
		app.viewModel.Tiles().CeilingTexture().Selected().Set(ceilingTexture.Value().(string))
		app.viewModel.Tiles().WallTexture().Selected().Set(wallTexture.Value().(string))
		app.viewModel.Tiles().FloorTextureRotations().Selected().Set(floorTextureRotations.Value().(string))
		app.viewModel.Tiles().CeilingTextureRotations().Selected().Set(ceilingTextureRotations.Value().(string))
		app.viewModel.Tiles().UseAdjacentWallTexture().Selected().Set(useAdjacentWallTexture.Value().(string))
		app.viewModel.Tiles().WallTextureOffset().Selected().Set(wallTextureOffset.Value().(string))
	})
}

func (app *MainApplication) onTileTypeChanged(newType string) {
	if newType != "" {
		app.requestSelectedTilesChange(func(properties *model.TileProperties) {
			properties.Type = new(model.TileType)
			*properties.Type = model.TileType(newType)
		}, true)
	}
}

func (app *MainApplication) onTileFloorHeightChanged(newValueString string) {
	newValue, err := strconv.ParseInt(newValueString, 10, 16)

	if newValueString != "" && err == nil {
		app.requestSelectedTilesChange(func(properties *model.TileProperties) {
			properties.FloorHeight = new(model.HeightUnit)
			*properties.FloorHeight = model.HeightUnit(int(newValue))
		}, true)
	}
}

func (app *MainApplication) onTileCeilingHeightChanged(newValueString string) {
	newValue, err := strconv.ParseInt(newValueString, 10, 16)

	if newValueString != "" && err == nil {
		app.requestSelectedTilesChange(func(properties *model.TileProperties) {
			properties.CeilingHeight = new(model.HeightUnit)
			*properties.CeilingHeight = model.HeightUnit(32 - int(newValue))
		}, true)
	}
}

func (app *MainApplication) onTileSlopeHeightChanged(newValueString string) {
	newValue, err := strconv.ParseInt(newValueString, 10, 16)

	if newValueString != "" && err == nil {
		app.requestSelectedTilesChange(func(properties *model.TileProperties) {
			properties.SlopeHeight = new(model.HeightUnit)
			*properties.SlopeHeight = model.HeightUnit(int(newValue))
		}, true)
	}
}

func (app *MainApplication) onTileSlopeControlChanged(newValue string) {
	if newValue != "" {
		app.requestSelectedTilesChange(func(properties *model.TileProperties) {
			properties.SlopeControl = new(model.SlopeControl)
			*properties.SlopeControl = model.SlopeControl(newValue)
		}, true)
	}
}

func (app *MainApplication) tileIntValueChangeCallback(accessor func(*model.TileProperties) **int, updateNeighbors bool) func(string) {
	return func(newValueString string) {
		newValue, err := strconv.ParseInt(newValueString, 10, 16)

		if newValueString != "" && err == nil {
			app.requestSelectedTilesChange(func(properties *model.TileProperties) {
				intPointer := accessor(properties)
				*intPointer = new(int)
				**intPointer = int(newValue)
			}, updateNeighbors)
		}
	}
}

func (app *MainApplication) tileIntRealWorldValueChangeCallback(accessor func(*model.RealWorldTileProperties) **int,
	updateNeighbors bool) func(string) {
	return app.tileIntValueChangeCallback(func(properties *model.TileProperties) **int {
		if properties.RealWorld == nil {
			properties.RealWorld = &model.RealWorldTileProperties{}
		}
		return accessor(properties.RealWorld)
	}, updateNeighbors)
}

func (app *MainApplication) onTileWallTextureOffsetChanged(newValueString string) {
	newValue, err := strconv.ParseInt(newValueString, 10, 16)

	if newValueString != "" && err == nil {
		app.requestSelectedTilesChange(func(properties *model.TileProperties) {
			if properties.RealWorld == nil {
				properties.RealWorld = &model.RealWorldTileProperties{}
			}
			properties.RealWorld.WallTextureOffset = new(model.HeightUnit)
			*properties.RealWorld.WallTextureOffset = model.HeightUnit(int(newValue))
		}, true)
	}
}

func (app *MainApplication) onTileUseAdjacentWallTextureChanged(newValue string) {
	if newValue != "" {
		app.requestSelectedTilesChange(func(properties *model.TileProperties) {
			if properties.RealWorld == nil {
				properties.RealWorld = &model.RealWorldTileProperties{}
			}
			properties.RealWorld.UseAdjacentWallTexture = new(bool)
			*properties.RealWorld.UseAdjacentWallTexture = newValue == "yes"
		}, true)
	}
}

func (app *MainApplication) onLevelTextureIndexChanged(newValueString string) {
	newValue, err := strconv.ParseInt(newValueString, 10, 16)

	app.updateViewModel(func() {
		if (newValueString != "") && (err == nil) && (newValue >= 0) && (int(newValue) < len(app.levelTextures)) {
			app.viewModel.LevelTextureID().Selected().Set(fmt.Sprintf("%d", app.levelTextures[newValue]))
		} else {
			app.viewModel.LevelTextureID().Selected().Set("")
		}
	})
}

func (app *MainApplication) onLevelTextureIDChanged(newValueString string) {
	if !app.viewModelUpdating {
		newValue, idErr := strconv.ParseInt(newValueString, 10, 16)
		index, indexErr := strconv.ParseInt(app.viewModel.LevelTextureIndex().Selected().Get(), 10, 16)

		if (newValueString != "") && (idErr == nil) && (indexErr == nil) &&
			(newValue >= 0) && (int(newValue) < len(app.textureData)) {
			app.levelTextures[index] = int(newValue)

			projectID := app.viewModel.SelectedProject()
			archiveID := "archive"
			levelID := app.activeLevelID
			app.store.SetLevelTextures(projectID, archiveID, levelID, app.levelTextures,
				app.onStoreLevelTexturesChanged, app.simpleStoreFailure("SetLevelTextures"))
		}
	}
}

func (app *MainApplication) onStoreLevelObjectsChanged(levelObjects *model.LevelObjects) {
	objectCount := len(levelObjects.Table)
	app.levelObjects = make(map[int]*editormodel.LevelObject)
	app.levelObjectIndices = make([]int, objectCount)
	for i := 0; i < objectCount; i++ {
		data := &levelObjects.Table[i]
		id := editormodel.MakeObjectID(data.Class, data.Subclass, data.Type)
		levelObject := editormodel.NewLevelObject(data, app.iconRetriever(id))
		app.levelObjects[levelObject.Index()] = levelObject
		app.levelObjectIndices[i] = levelObject.Index()
	}
	app.viewModel.LevelObjects().SetObjectCount(objectCount)
}

func (app *MainApplication) iconRetriever(id editormodel.ObjectID) graphics.BitmapRetriever {
	if app.gameObjectIconRetriever[id] == nil {
		projectID := app.viewModel.SelectedProject()

		app.gameObjectIconRetriever[id] = func() *graphics.BitmapTexture { return app.defaultIcon }

		app.store.GameObjectIcon(projectID, id.Class(), id.Subclass(), id.Type(),
			func(raw *model.RawBitmap) {
				bmp := graphics.BitmapFromRaw(*raw)
				bitmap := graphics.NewBitmapTexture(app.gl, bmp.Width, bmp.Height, bmp.Pixels)
				app.gameObjectIcons[id] = bitmap
				app.gameObjectIconRetriever[id] = func() *graphics.BitmapTexture { return bitmap }
			}, app.simpleStoreFailure("GameObjectIcon"))
	}

	return func() *graphics.BitmapTexture { return app.gameObjectIconRetriever[id]() }
}

func (app *MainApplication) onSelectedObjectIndexChanged(string) {
	index := app.viewModel.LevelObjects().SelectedObjectIndex()
	levelObject := app.levelObjects[index]
	class, subclass, objType := -1, -1, -1
	rawDataString := ""
	var propertyNodes []viewmodel.Node

	if levelObject != nil {
		var classDataBytes []byte
		class, subclass, objType = levelObject.ID()
		for index, b := range levelObject.ClassData() {
			if index > 0 {
				rawDataString += " "
			}
			rawDataString += fmt.Sprintf("%02X", b)
			classDataBytes = append(classDataBytes, byte(b))
		}

		objID := res.MakeObjectID(res.ObjectClass(class), res.ObjectSubclass(subclass), res.ObjectType(objType))
		var rootInterpreter *interpreters.Instance

		if app.isActiveLevelRealWorld() {
			rootInterpreter = levelobj.ForRealWorld(objID, classDataBytes)
		} else {
			rootInterpreter = levelobj.ForCyberspace(objID, classDataBytes)
		}

		var listProperties func(path string, interpreter *interpreters.Instance)

		listProperties = func(path string, interpreter *interpreters.Instance) {
			propertyKeys := interpreter.Keys()
			for _, propertyKey := range propertyKeys {
				propertyNodes = append(propertyNodes, viewmodel.NewStringValueNode(path+propertyKey, fmt.Sprintf("%v", interpreter.Get(propertyKey))))
			}
			refinementKeys := interpreter.ActiveRefinements()
			for _, refinementKey := range refinementKeys {
				listProperties(path+"."+refinementKey, interpreter.Refined(refinementKey))
			}
		}
		listProperties("", rootInterpreter)
	}
	app.viewModel.LevelObjects().SetObjectID(class, subclass, objType)
	app.viewModel.LevelObjects().RawData().Set(rawDataString)
	app.viewModel.LevelObjects().Properties().Set(propertyNodes)
}

package editor

import (
	"encoding/base64"
	"fmt"
	"os"

	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/editor/camera"
	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/opengl"
	"github.com/inkyblackness/shocked-model"
)

// MainApplication represents the core intelligence of the editor.
type MainApplication struct {
	store DataStore

	glWindow env.OpenGlWindow
	gl       opengl.OpenGl

	mouseX, mouseY   float32
	mouseMoveCapture func()

	view *camera.LimitedCamera

	gridRenderable           *GridRenderable
	textureRenderables       []*TextureRenderable
	tileTextureMapRenderable *TileTextureMapRenderable
}

// NewMainApplication returns a new instance of MainApplication.
func NewMainApplication(store DataStore) *MainApplication {
	camLimit := (TilesPerMapSide - 1) * TileBaseLength

	return &MainApplication{
		store:            store,
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
	app.gl.BlendFunc(opengl.SRC_ALPHA, opengl.ONE_MINUS_SRC_ALPHA)
	app.gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	app.gridRenderable = NewGridRenderable(app.gl)

	/*
		if app.store != nil {
			app.store.Palette("test1", "game", func(colors [256]model.Color) {
				fmt.Fprintf(os.Stderr, "!!!!! palette: %v\n", colors)
			}, app.simpleStoreFailure("Palette"))
			app.store.LevelTextures("test1", "archive", 1, func(textureIDs []int) {
				fmt.Fprintf(os.Stderr, "!!!!! Level Textures: %v\n", textureIDs)
				for _, id := range textureIDs {
					app.store.TextureBitmap("test1", id, "icon", func(bmp *model.RawBitmap) {
						fmt.Fprintf(os.Stderr, "!!!!! bitmap: %v\n", bmp)
					}, app.simpleStoreFailure("TextureBitmap"))
				}
			}, app.simpleStoreFailure("LevelTextures"))
		}
	*/

	/* load a single texture
	if app.store != nil {
		var paletteTexture GraphicsTexture
		var bitmapTexture GraphicsTexture

		createTextureRenderable := func() {
			if paletteTexture != nil && bitmapTexture != nil && len(app.textureRenderables) == 0 {
				app.textureRenderables = append(app.textureRenderables, NewTextureRenderable(app.gl,
					64.0, 128.0, 256.0,
					paletteTexture, bitmapTexture))
			}
		}

		app.store.LevelTextures("test1", "archive", 1, func(textureIDs []int) {
			app.store.TextureBitmap("test1", textureIDs[7], "large", func(bmp *model.RawBitmap) {
				pixelData, _ := base64.StdEncoding.DecodeString(bmp.Pixel)
				bitmapTexture = NewBitmapTexture(app.gl, bmp.Width, bmp.Height, pixelData)
				createTextureRenderable()
			}, app.simpleStoreFailure("TextureBitmap"))
		}, app.simpleStoreFailure("LevelTextures"))

		app.store.Palette("test1", "game", func(colors [256]model.Color) {
			colorProvider := func(index int) (byte, byte, byte, byte) {
				entry := &colors[index]
				return byte(entry.Red), byte(entry.Green), byte(entry.Blue), 255
			}
			paletteTexture = NewPaletteTexture(app.gl, colorProvider)

			createTextureRenderable()
		}, app.simpleStoreFailure("Palette"))
	}
	*/

	if app.store != nil {
		var paletteTexture GraphicsTexture
		var levelTextureIDs []int
		bitmapTextures := make(map[int]GraphicsTexture)
		var tiles *model.Tiles

		createMap := func() {
			if paletteTexture != nil && tiles != nil &&
				len(levelTextureIDs) > 0 && len(bitmapTextures) == len(levelTextureIDs) &&
				app.tileTextureMapRenderable == nil {

				fmt.Fprintf(os.Stderr, "creating tile map\n")
				app.tileTextureMapRenderable = NewTileTextureMapRenderable(app.gl, paletteTexture)

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

		app.store.Tiles("test1", "archive", 1, func(data model.Tiles) {
			fmt.Fprintf(os.Stderr, "loaded tiles\n")
			tiles = &data

			createMap()
		}, app.simpleStoreFailure("Tiles"))

		app.store.LevelTextures("test1", "archive", 1, func(textureIDs []int) {
			fmt.Fprintf(os.Stderr, "loaded textureIDs, %v to load\n", len(textureIDs))
			levelTextureIDs = textureIDs
			for index, id := range textureIDs {
				localIndex := index
				app.store.TextureBitmap("test1", id, "large", func(bmp *model.RawBitmap) {
					pixelData, _ := base64.StdEncoding.DecodeString(bmp.Pixel)
					bitmapTextures[localIndex] = NewBitmapTexture(app.gl, bmp.Width, bmp.Height, pixelData)

					createMap()
				}, app.simpleStoreFailure("TextureBitmap"))
			}
		}, app.simpleStoreFailure("LevelTextures"))

		app.store.Palette("test1", "game", func(colors [256]model.Color) {
			colorProvider := func(index int) (byte, byte, byte, byte) {
				entry := &colors[index]
				return byte(entry.Red), byte(entry.Green), byte(entry.Blue), 255
			}
			paletteTexture = NewPaletteTexture(app.gl, colorProvider)

			createMap()
		}, app.simpleStoreFailure("Palette"))
	}

	/* load full map - slow
	if app.store != nil {
		var palette *[256]model.Color
		var levelTextureIDs []int
		bitmaps := make(map[int]*model.RawBitmap)
		bitmapData := make(map[int][]byte)
		var tiles *model.Tiles

		createMap := func() {
			fmt.Fprintf(os.Stderr, "creating map? got %v bitmaps, palette: %v, tiles: %v, renderables: %v\n",
				len(bitmaps), palette != nil, tiles != nil, len(app.textureRenderables))
			if palette != nil && tiles != nil && len(levelTextureIDs) > 0 &&
				len(bitmaps) == len(levelTextureIDs) && len(app.textureRenderables) == 0 {
				fmt.Fprintf(os.Stderr, "creating map!\n")

				colorProvider := func(index int) (byte, byte, byte, byte) {
					entry := &palette[index]
					return byte(entry.Red), byte(entry.Green), byte(entry.Blue), 255
				}

				for y := 0; y < len(tiles.Table); y++ {
					row := tiles.Table[y]
					for x := 0; x < len(row); x++ {
						tileData := &row[x]
						textureID := *tileData.Properties.RealWorld.FloorTexture

						bitmap := bitmaps[textureID]
						pixelData := bitmapData[textureID]
						renderable := NewTextureRenderable(app.gl,
							float32(63-y)*32.0, float32(x)*32.0, 32.0,
							bitmap.Width, bitmap.Height, pixelData, colorProvider)
						app.textureRenderables = append(app.textureRenderables, renderable)
					}
				}
			}
		}

		app.store.Tiles("test1", "archive", 1, func(data model.Tiles) {
			fmt.Fprintf(os.Stderr, "loaded tiles\n")
			tiles = &data
			createMap()
		}, app.simpleStoreFailure("Tiles"))

		app.store.LevelTextures("test1", "archive", 1, func(textureIDs []int) {
			fmt.Fprintf(os.Stderr, "loaded textureIDs, %v to load\n", len(textureIDs))
			levelTextureIDs = textureIDs
			for index, id := range textureIDs {
				localIndex := index
				app.store.TextureBitmap("test1", id, "large", func(bmp *model.RawBitmap) {
					pixelData, _ := base64.StdEncoding.DecodeString(bmp.Pixel)
					bitmapData[localIndex] = pixelData
					bitmaps[localIndex] = bmp
					createMap()
				}, app.simpleStoreFailure("TextureBitmap"))
			}
		}, app.simpleStoreFailure("LevelTextures"))

		app.store.Palette("test1", "game", func(colors [256]model.Color) {
			fmt.Fprintf(os.Stderr, "loaded palete\n")
			palette = &colors
			createMap()
		}, app.simpleStoreFailure("Palette"))
	}
	*/
}

func (app *MainApplication) simpleStoreFailure(info string) FailureFunc {
	return func() {
		fmt.Fprintf(os.Stderr, "Failed to process store query <%s>\n", info)
	}
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
	for _, renderable := range app.textureRenderables {
		renderable.Render(&context)
	}
	if app.tileTextureMapRenderable != nil {
		app.tileTextureMapRenderable.Render(&context)
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

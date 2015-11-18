package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Archs/js/JSON"
	"github.com/Archs/js/gopherjs-ko"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/inkyblackness/shocked-model"
)

func getResource(url string, responseData interface{}, onSuccess func(), onFailure func()) {
	ajaxopt := map[string]interface{}{
		"method":   "GET",
		"url":      url,
		"dataType": "json",
		"data":     nil,
		"jsonp":    false,
		"success": func(data interface{}) {
			dataString := JSON.Stringify(data)

			json.Unmarshal(bytes.NewBufferString(dataString).Bytes(), responseData)
			onSuccess()
		},
		"error": func(status interface{}) {
			onFailure()
		},
	}

	jquery.Ajax(ajaxopt)
}

type ViewModel struct {
	*ko.BaseViewModel
	MapWidth  *ko.Observable `js:"mapWidth"`
	MapHeight *ko.Observable `js:"mapHeight"`
	TileRows  *ko.Observable `js:"tileRows"`

	Levels        *ko.Observable `js:"levels"`
	SelectedLevel *ko.Observable `js:"selectedLevel"`

	LevelTextures *ko.Observable `js:"levelTextures"`

	ShouldShowFloorTexture   *ko.Observable `js:"shouldShowFloorTexture"`
	ShouldShowCeilingTexture *ko.Observable `js:"shouldShowCeilingTexture"`
}

type TileRow struct {
	*ko.BaseViewModel
	Y           int            `js:"y"`
	TileColumns *ko.Observable `js:"tileColumns"`
}

type Tile struct {
	*ko.BaseViewModel
	X int `js:"x"`

	FloorTextureIndex     *ko.Observable `js:"floorTextureIndex"`
	FloorTextureRotations *ko.Observable `js:"floorTextureRotations"`
	FloorTextureUrl       *ko.Observable `js:"floorTextureUrl"`

	CeilingTextureIndex     *ko.Observable `js:"ceilingTextureIndex"`
	CeilingTextureRotations *ko.Observable `js:"ceilingTextureRotations"`
	CeilingTextureUrl       *ko.Observable `js:"ceilingTextureUrl"`

	TileType *ko.Observable `js:"tileType"`
}

type Level struct {
	*ko.BaseViewModel
	ID int `js:"id"`

	IsSelected *ko.Observable `js:"isSelected"`

	Select func() `js:"select"`
}

func New() *ViewModel {
	self := new(ViewModel)
	self.BaseViewModel = ko.NewBaseViewModel()
	self.MapWidth = ko.NewObservable(0)
	self.MapHeight = ko.NewObservable(0)
	self.TileRows = ko.NewObservableArray()
	self.Levels = ko.NewObservableArray()
	self.SelectedLevel = ko.NewObservable(-1)
	self.LevelTextures = ko.NewObservableArray()
	self.LevelTextures.RateLimit(500, true)

	self.ShouldShowFloorTexture = ko.NewObservable(true)
	self.ShouldShowCeilingTexture = ko.NewObservable(false)
	return self
}

func main() {
	vm := New()

	resizeTileColumns := func(y int, list *ko.Observable, newWidth int) {
		for list.Length() > newWidth {
			list.Pop()
		}
		for list.Length() < newWidth {
			x := list.Length()
			tile := new(Tile)

			tile.BaseViewModel = ko.NewBaseViewModel()
			tile.X = x
			tile.TileType = ko.NewObservable("solid")

			tile.FloorTextureIndex = ko.NewObservable(-1)
			tile.FloorTextureRotations = ko.NewObservable("rotations0")
			tile.FloorTextureUrl = ko.NewComputed(func() interface{} {
				textureIndex := tile.FloorTextureIndex.Get().Int()
				url := ""

				if (textureIndex >= 0) && (textureIndex < vm.LevelTextures.Length()) {
					url = fmt.Sprintf("/projects/test1/textures/%d/large/png", vm.LevelTextures.Index(textureIndex).Int())
				}

				return "url(" + url + ")"
			})
			tile.CeilingTextureIndex = ko.NewObservable(-1)
			tile.CeilingTextureRotations = ko.NewObservable("rotations0")
			tile.CeilingTextureUrl = ko.NewComputed(func() interface{} {
				textureIndex := tile.CeilingTextureIndex.Get().Int()
				url := ""

				if (textureIndex >= 0) && (textureIndex < vm.LevelTextures.Length()) {
					url = fmt.Sprintf("/projects/test1/textures/%d/large/png", vm.LevelTextures.Index(textureIndex).Int())
				}

				return "url(" + url + ")"
			})

			list.Push(tile)
		}
	}

	vm.MapWidth.Subscribe(func(obj *js.Object) {
		newWidth := obj.Int()

		for i := 0; i < vm.TileRows.Length(); i++ {
			tileRow := new(TileRow)
			tileRow.BaseViewModel = ko.NewBaseViewModel()
			tileRow.FromJS(vm.TileRows.Index(i))
			resizeTileColumns(i, tileRow.TileColumns, newWidth)
		}
	})

	vm.MapHeight.Subscribe(func(obj *js.Object) {
		newHeight := obj.Int()

		for vm.TileRows.Length() > newHeight {
			vm.TileRows.Pop()
		}
		for vm.TileRows.Length() < newHeight {
			y := newHeight - vm.TileRows.Length() - 1
			tileRow := new(TileRow)

			tileRow.BaseViewModel = ko.NewBaseViewModel()
			tileRow.Y = y
			tileRow.TileColumns = ko.NewObservableArray()
			resizeTileColumns(y, tileRow.TileColumns, vm.MapWidth.Get().Int())
			vm.TileRows.Push(tileRow)
		}
	})

	vm.MapWidth.Set(64)
	vm.MapHeight.Set(64)

	loadLevel := func(levelID int) {
		var levelTextures model.LevelTextures
		getResource(fmt.Sprintf("/projects/test1/archive/level/%d/textures", levelID), &levelTextures, func() {
			vm.LevelTextures.RemoveAll()
			for _, id := range levelTextures.IDs {
				vm.LevelTextures.Push(id)
			}
			println("textures found: ", vm.LevelTextures.Length())
		}, func() {})

		var tileMap model.Tiles
		getResource(fmt.Sprintf("/projects/test1/archive/level/%d/tiles", levelID), &tileMap, func() {
			for y, row := range tileMap.Table {
				for x, tileData := range row {
					var tileRow TileRow
					var tile Tile

					tile.BaseViewModel = ko.NewBaseViewModel()
					tileRow.BaseViewModel = ko.NewBaseViewModel()
					tileRow.FromJS(vm.TileRows.Index(vm.TileRows.Length() - 1 - y))
					tile.FromJS(tileRow.TileColumns.Index(x))
					tile.TileType.Set(tileData.Properties.Type)
					tile.FloorTextureIndex.Set(tileData.Properties.RealWorld.FloorTexture)
					tile.FloorTextureRotations.Set(fmt.Sprintf("rotations%d", tileData.Properties.RealWorld.FloorTextureRotations))
					tile.CeilingTextureIndex.Set(tileData.Properties.RealWorld.CeilingTexture)
					tile.CeilingTextureRotations.Set(fmt.Sprintf("rotations%d", tileData.Properties.RealWorld.CeilingTextureRotations))
				}
			}
		}, func() {})
	}

	selectLevel := func(levelID int) func() {
		return func() {
			if vm.SelectedLevel.Get().Int() != levelID {
				vm.SelectedLevel.Set(levelID)
				loadLevel(levelID)
			}
		}
	}

	for levelID := 0; levelID < 16; levelID++ {
		level := new(Level)
		level.BaseViewModel = ko.NewBaseViewModel()
		level.ID = levelID
		level.IsSelected = ko.NewComputed(func() interface{} {
			return vm.SelectedLevel.Get().Int() == levelID
		})
		level.Select = selectLevel(levelID)
		vm.Levels.Push(level)
	}

	ko.ApplyBindings(vm)

}

package model

import (
	"strconv"

	"github.com/inkyblackness/shocked-model"
)

// LevelAdapter is the entry point for a level.
type LevelAdapter struct {
	context            archiveContext
	store              model.DataStore
	simpleStoreFailure func(string) model.FailureFunc

	id           *observable
	isCyberspace bool
	tileMap      *TileMap
}

func newLevelAdapter(context archiveContext, store model.DataStore, simpleStoreFailure func(string) model.FailureFunc) *LevelAdapter {
	adapter := &LevelAdapter{
		context:            context,
		store:              store,
		simpleStoreFailure: simpleStoreFailure,

		id:      newObservable(),
		tileMap: NewTileMap(64, 64)}

	adapter.id.set("")

	return adapter
}

// ID returns the ID of the level.
func (adapter *LevelAdapter) ID() string {
	return adapter.id.orDefault("").(string)
}

func (adapter *LevelAdapter) storeLevelID() int {
	idAsString := adapter.ID()
	id := -1

	if idAsString != "" {
		parsed, _ := strconv.ParseInt(idAsString, 10, 16)
		id = int(parsed)
	}

	return id
}

// OnIDChanged registers a callback for changed IDs.
func (adapter *LevelAdapter) OnIDChanged(callback func()) {
	adapter.id.addObserver(callback)
}

func (adapter *LevelAdapter) requestByID(levelID string) {
	adapter.id.set("")
	adapter.tileMap.clear()

	adapter.id.set(levelID)
	if levelID != "" {
		adapter.store.Tiles(adapter.context.ActiveProjectID(), adapter.context.ActiveArchiveID(), adapter.storeLevelID(),
			adapter.onTiles, adapter.simpleStoreFailure("Tiles"))
	}
}

// IsCyberspace returns true for cyberspace levels.
func (adapter *LevelAdapter) IsCyberspace() bool {
	return adapter.isCyberspace
}

// TileMap returns the map of tiles of the level.
func (adapter *LevelAdapter) TileMap() *TileMap {
	return adapter.tileMap
}

func (adapter *LevelAdapter) onTiles(tiles model.Tiles) {
	for y := 0; y < len(tiles.Table); y++ {
		row := tiles.Table[y]
		for x := 0; x < len(row); x++ {
			tileProperties := &row[x].Properties
			coord := TileCoordinateOf(x, y)
			tile := adapter.tileMap.Tile(coord)
			tile.setProperties(tileProperties)
		}
	}
}

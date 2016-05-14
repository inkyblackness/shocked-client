package model

// TileMap contains the tiles with their properties.
type TileMap struct {
	tiles         map[TileCoordinate]*Tile
	selectedTiles map[TileCoordinate]*Tile
}

// NewTileMap returns a new tile map instance
func NewTileMap(width, height int) *TileMap {
	tileMap := &TileMap{
		tiles:         make(map[TileCoordinate]*Tile),
		selectedTiles: make(map[TileCoordinate]*Tile)}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			coord := TileCoordinateOf(x, y)
			tileMap.tiles[coord] = NewTile()
		}
	}

	return tileMap
}

// ForEachSelected iterates through all selected tiles and calls the specified callback.
func (tileMap *TileMap) ForEachSelected(callback func(coord TileCoordinate, tile *Tile)) {
	for coord, tile := range tileMap.selectedTiles {
		callback(coord, tile)
	}
}

// ClearSelection clears the current selection.
func (tileMap *TileMap) ClearSelection() {
	tileMap.selectedTiles = make(map[TileCoordinate]*Tile)
}

// SetSelected sets the selection state of the tile at given coordinate.
func (tileMap *TileMap) SetSelected(coord TileCoordinate, value bool) {
	_, isSelected := tileMap.selectedTiles[coord]
	tile, exists := tileMap.tiles[coord]

	if isSelected && !value {
		delete(tileMap.selectedTiles, coord)
	} else if !isSelected && value && exists {
		tileMap.selectedTiles[coord] = tile
	}
}
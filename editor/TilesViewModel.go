package editor

import (
	"github.com/inkyblackness/shocked-client/viewmodel"

	"github.com/inkyblackness/shocked-model"
)

// TilesViewModel contains the view model entries for the map tiles.
type TilesViewModel struct {
	root *viewmodel.SectionNode

	tileType *viewmodel.ValueSelectionNode
}

// NewTilesViewModel returns a new instance of a TilesViewModel.
func NewTilesViewModel() *TilesViewModel {
	vm := &TilesViewModel{}

	vm.tileType = viewmodel.NewValueSelectionNode("Tile Type", []string{string(model.Open), string(model.Solid),
		string(model.DiagonalOpenSouthEast), string(model.DiagonalOpenSouthWest), string(model.DiagonalOpenNorthWest), string(model.DiagonalOpenNorthEast),
		string(model.SlopeSouthToNorth), string(model.SlopeWestToEast), string(model.SlopeNorthToSouth), string(model.SlopeEastToWest),
		string(model.ValleySouthEastToNorthWest), string(model.ValleySouthWestToNorthEast), string(model.ValleyNorthWestToSouthEast), string(model.ValleyNorthEastToSouthWest),
		string(model.RidgeNorthWestToSouthEast), string(model.RidgeNorthEastToSouthWest), string(model.RidgeSouthEastToNorthWest), string(model.RidgeSouthWestToNorthEast),
		""},
		"")

	vm.root = viewmodel.NewSectionNode("Tiles",
		[]viewmodel.Node{vm.tileType},
		viewmodel.NewBoolValueNode("", true))

	return vm
}

// TileType returns the tile type selection node.
func (vm *TilesViewModel) TileType() *viewmodel.ValueSelectionNode {
	return vm.tileType
}

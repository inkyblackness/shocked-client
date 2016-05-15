package editor

import (
	"fmt"

	"github.com/inkyblackness/shocked-client/viewmodel"
	"github.com/inkyblackness/shocked-model"
)

// TilesViewModel contains the view model entries for the map tiles.
type TilesViewModel struct {
	root *viewmodel.SectionNode

	tileType      *viewmodel.ValueSelectionNode
	floorHeight   *viewmodel.ValueSelectionNode
	ceilingHeight *viewmodel.ValueSelectionNode
	slopeHeight   *viewmodel.ValueSelectionNode
	slopeControl  *viewmodel.ValueSelectionNode
}

func heightsForSelection(start, stop int) (list []string) {
	for level := start; level <= stop; level++ {
		list = append(list, fmt.Sprintf("%d", level))
	}

	return append(list, "")
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
	vm.floorHeight = viewmodel.NewValueSelectionNode("Floor Height Level", heightsForSelection(0, 31), "")
	vm.ceilingHeight = viewmodel.NewValueSelectionNode("Ceiling Height Level", heightsForSelection(1, 32), "")
	vm.slopeHeight = viewmodel.NewValueSelectionNode("Slope Height", heightsForSelection(0, 31), "")
	vm.slopeControl = viewmodel.NewValueSelectionNode("Slope Control",
		[]string{model.SlopeCeilingInverted, model.SlopeCeilingMirrored, model.SlopeCeilingFlat, model.SlopeFloorFlat, ""},
		"")

	vm.root = viewmodel.NewSectionNode("Tiles",
		[]viewmodel.Node{vm.tileType, vm.floorHeight, vm.ceilingHeight, vm.slopeHeight, vm.slopeControl},
		viewmodel.NewBoolValueNode("", true))

	return vm
}

// TileType returns the tile type selection node.
func (vm *TilesViewModel) TileType() *viewmodel.ValueSelectionNode {
	return vm.tileType
}

// FloorHeight returns the floor height selection node.
func (vm *TilesViewModel) FloorHeight() *viewmodel.ValueSelectionNode {
	return vm.floorHeight
}

// CeilingHeight returns the ceiling height selection node.
func (vm *TilesViewModel) CeilingHeight() *viewmodel.ValueSelectionNode {
	return vm.ceilingHeight
}

// SlopeHeight returns the slope height selection node.
func (vm *TilesViewModel) SlopeHeight() *viewmodel.ValueSelectionNode {
	return vm.slopeHeight
}

// SlopeControl returns the slope control selection node.
func (vm *TilesViewModel) SlopeControl() *viewmodel.ValueSelectionNode {
	return vm.slopeControl
}

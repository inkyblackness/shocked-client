package editor

import (
	"fmt"

	"github.com/inkyblackness/shocked-client/viewmodel"
)

// ViewModel contains the raw view model node structure, wrapped with simple accessors.
type ViewModel struct {
	root *viewmodel.SectionNode

	mainSection *viewmodel.SectionSelectionNode

	projects     *viewmodel.ValueSelectionNode
	textureCount *viewmodel.StringValueNode
	levels       *viewmodel.ValueSelectionNode

	pointerCoordinate *viewmodel.StringValueNode

	tiles *TilesViewModel
}

// NewViewModel returns a new ViewModel instance.
func NewViewModel() *ViewModel {
	vm := &ViewModel{}

	vm.projects = viewmodel.NewValueSelectionNode("Select", nil, "")
	vm.textureCount = viewmodel.NewStringValueNode("Texture Count", "")
	projectSection := viewmodel.NewSectionNode("Project", []viewmodel.Node{vm.projects, vm.textureCount},
		viewmodel.NewBoolValueNode("Available", true))

	vm.levels = viewmodel.NewValueSelectionNode("Level", nil, "")
	vm.tiles = NewTilesViewModel()
	mapControlSection := viewmodel.NewSectionNode("Control", []viewmodel.Node{vm.levels}, viewmodel.NewBoolValueNode("", true))
	mapSectionSelection := viewmodel.NewSectionSelectionNode("Map Section", map[string]*viewmodel.SectionNode{
		"Control": mapControlSection,
		"Tiles":   vm.tiles.root}, "Control")

	projectSelected := viewmodel.NewBoolValueNode("Available", false)
	vm.projects.Selected().Subscribe(func(projectID string) {
		projectSelected.Set(projectID != "")
	})
	mapSection := viewmodel.NewSectionNode("Map", []viewmodel.Node{mapSectionSelection}, projectSelected)

	vm.mainSection = viewmodel.NewSectionSelectionNode("Section", map[string]*viewmodel.SectionNode{
		"Project": projectSection,
		"Map":     mapSection}, "Project")

	vm.pointerCoordinate = viewmodel.NewStringValueNode("Pointer at", "")

	vm.root = viewmodel.NewSectionNode("",
		[]viewmodel.Node{vm.mainSection, vm.pointerCoordinate},
		viewmodel.NewBoolValueNode("", true))

	return vm
}

// Root returns the entry point to the raw node structure.
func (vm *ViewModel) Root() viewmodel.Node {
	return vm.root
}

// SelectMapSection ensures the map controls are selected.
func (vm *ViewModel) SelectMapSection() {
	vm.mainSection.Selection().Selected().Set("Map")
}

// SelectedProject returns the identifier of the currently selected project.
func (vm *ViewModel) SelectedProject() string {
	return vm.projects.Selected().Get()
}

// OnSelectedProjectChanged registers a callback for a change in the selected project
func (vm *ViewModel) OnSelectedProjectChanged(callback func(projectID string)) {
	vm.projects.Selected().Subscribe(callback)
}

// SetProjects sets the list of available project identifier.
func (vm *ViewModel) SetProjects(projectIDs []string) {
	vm.projects.SetValues(projectIDs)
}

// SelectProject sets the currently selected project.
func (vm *ViewModel) SelectProject(id string) {
	vm.projects.Selected().Set(id)
}

// SetTextureCount sets the amount of textures of the project.
func (vm *ViewModel) SetTextureCount(value int) {
	vm.textureCount.Set(fmt.Sprintf("%d", value))
}

// OnSelectedLevelChanged registers a callback for a change in the selected level
func (vm *ViewModel) OnSelectedLevelChanged(callback func(levelID string)) {
	vm.levels.Selected().Subscribe(callback)
}

// SetLevels sets the list of available level identifier.
func (vm *ViewModel) SetLevels(levelIDs []string) {
	vm.levels.SetValues(levelIDs)
}

// SetPointerAt registers where the pointer is currently hovering at.
func (vm *ViewModel) SetPointerAt(tileX, tileY int, subX, subY int) {
	text := ""

	if (tileX >= 0) && (tileY >= 0) && (tileX < int(TilesPerMapSide)) && (tileY < int(TilesPerMapSide)) {
		text = fmt.Sprintf("Tile: %2d/%2d Sub: %3d/%3d", tileX, tileY, subX, subY)
	}
	vm.pointerCoordinate.Set(text)
}

// Tiles returns the sub-section about tiles.
func (vm *ViewModel) Tiles() *TilesViewModel {
	return vm.tiles
}

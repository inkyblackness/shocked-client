package editor

import (
	"github.com/inkyblackness/shocked-client/viewmodel"
)

// ViewModel contains the raw view model node structure, wrapped with simple accessors.
type ViewModel struct {
	root viewmodel.Node

	projects *viewmodel.ValueSelectionNode
}

// NewViewModel returns a new ViewModel instance.
func NewViewModel() *ViewModel {
	vm := &ViewModel{}

	vm.projects = viewmodel.NewValueSelectionNode("Select", nil, "")
	projectSection := viewmodel.NewSectionNode("Project", []viewmodel.Node{vm.projects}, viewmodel.NewBoolValueNode("Available", true))

	mapSectionAvailable := viewmodel.NewBoolValueNode("Available", false)
	vm.projects.Selected().Subscribe(func(projectID string) {
		mapSectionAvailable.Set(projectID != "")
	})
	mapSection := viewmodel.NewSectionNode("Map", []viewmodel.Node{}, mapSectionAvailable)

	vm.root = viewmodel.NewSectionSelectionNode("Section", map[string]*viewmodel.SectionNode{
		"Project": projectSection,
		"Map":     mapSection}, "Project")

	return vm
}

// Root returns the entry point to the raw node structure.
func (vm *ViewModel) Root() viewmodel.Node {
	return vm.root
}

// OnSelectedProjectChanged registers a callback for a change in the selected project
func (vm *ViewModel) OnSelectedProjectChanged(callback func(projectID string)) {
	vm.projects.Selected().Subscribe(callback)
}

// SetProjects sets the list of available project identifier.
func (vm *ViewModel) SetProjects(projectIDs []string) {
	vm.projects.SetValues(projectIDs)
}

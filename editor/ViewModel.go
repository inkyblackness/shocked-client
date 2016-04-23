package editor

import (
	"github.com/inkyblackness/shocked-client/viewmodel"
)

// ViewModel contains the raw view model node structure, wrapped with simple accessors.
type ViewModel struct {
	root *viewmodel.ContainerNode

	selectedProject   *viewmodel.StringValueNode
	availableProjects *viewmodel.ArrayNode
}

// NewViewModel returns a new ViewModel instance.
func NewViewModel() *ViewModel {
	vm := &ViewModel{}

	view := viewmodel.NewContainerNode(map[string]viewmodel.Node{
		"selectedMainSection": viewmodel.NewStringValueNode("project"),
		"mainSections": viewmodel.NewArrayNode(
			viewmodel.NewStringValueNode("project"),
			viewmodel.NewStringValueNode("main"),
			viewmodel.NewStringValueNode("empty"))})

	vm.selectedProject = viewmodel.NewStringValueNode("")
	vm.availableProjects = viewmodel.NewArrayNode()
	projects := viewmodel.NewContainerNode(map[string]viewmodel.Node{
		"selected":  vm.selectedProject,
		"available": vm.availableProjects})

	vm.root = viewmodel.NewContainerNode(map[string]viewmodel.Node{
		"view":     view,
		"projects": projects})

	return vm
}

// Root returns the entry point to the raw node structure.
func (vm *ViewModel) Root() viewmodel.Node {
	return vm.root
}

// OnSelectedProjectChanged registers a callback for a change in the selected project
func (vm *ViewModel) OnSelectedProjectChanged(callback func(projectID string)) {
	vm.selectedProject.Subscribe(callback)
}

// SetProjects sets the list of available project identifier.
func (vm *ViewModel) SetProjects(projectIDs []string) {
	nodes := make([]viewmodel.Node, len(projectIDs))
	for index, id := range projectIDs {
		nodes[index] = viewmodel.NewStringValueNode(id)
	}
	vm.availableProjects.Set(nodes)
}

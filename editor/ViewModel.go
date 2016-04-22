package editor

import (
	"github.com/inkyblackness/shocked-client/viewmodel"
)

// ViewModel contains the raw view model node structure, wrapped with simple accessors.
type ViewModel struct {
	root *viewmodel.ContainerNode
}

// NewViewModel returns a new ViewModel instance.
func NewViewModel() *ViewModel {
	view := viewmodel.NewContainerNode(map[string]viewmodel.Node{
		"selectedMainSection": viewmodel.NewStringValueNode("main"),
		"mainSections": viewmodel.NewArrayNode(
			viewmodel.NewStringValueNode("main"),
			viewmodel.NewStringValueNode("empty"))})

	projects := viewmodel.NewContainerNode(map[string]viewmodel.Node{
		"active":    viewmodel.NewStringValueNode(""),
		"available": viewmodel.NewArrayNode()})

	root := viewmodel.NewContainerNode(map[string]viewmodel.Node{
		"view":     view,
		"projects": projects})

	vm := &ViewModel{root}

	return vm
}

// Root returns the entry point to the raw node structure.
func (vm *ViewModel) Root() viewmodel.Node {
	return vm.root
}

package editor

import (
	"fmt"
	"strconv"

	"github.com/inkyblackness/shocked-client/viewmodel"
)

// ObjectsViewModel contains the view model entries for the level objects.
type ObjectsViewModel struct {
	root *viewmodel.SectionNode

	selectedObject *viewmodel.ValueSelectionNode
	cst            *viewmodel.StringValueNode
	rawData        *viewmodel.StringValueNode
	properties     *viewmodel.SectionNode
}

// NewObjectsViewModel returns a new instance of a ObjectsViewModel.
func NewObjectsViewModel(levelIsRealWorld *viewmodel.BoolValueNode) *ObjectsViewModel {
	vm := &ObjectsViewModel{}

	vm.selectedObject = viewmodel.NewValueSelectionNode("Selected Object", []string{""}, "")
	vm.cst = viewmodel.NewStringValueNode("C/S/T", "")
	vm.rawData = viewmodel.NewEditableStringValueNode("Raw Data", "")
	vm.properties = viewmodel.NewSectionNode("Properties", nil, viewmodel.NewBoolValueNode("", true))

	vm.root = viewmodel.NewSectionNode("Objects",
		[]viewmodel.Node{vm.selectedObject, vm.cst, vm.rawData, vm.properties},
		viewmodel.NewBoolValueNode("", true))

	return vm
}

// SelectedObject returns the node for the object index selection.
func (vm *ObjectsViewModel) SelectedObject() *viewmodel.ValueSelectionNode {
	return vm.selectedObject
}

// SelectedObjectIndex returns the index of the currently selected object. -1 if none selected.
func (vm *ObjectsViewModel) SelectedObjectIndex() int {
	indexString := vm.selectedObject.Selected().Get()
	index, err := strconv.ParseInt(indexString, 10, 16)

	if err != nil {
		index = -1
	}
	return int(index)
}

// SetObjectCount registers the available amount of level objects.
func (vm *ObjectsViewModel) SetObjectCount(count int) {
	values := []string{""}

	if count > 0 {
		values = intStringList(0, count-1)
	}
	vm.selectedObject.SetValues(values)
}

// SetObjectID sets the identification value
func (vm *ObjectsViewModel) SetObjectID(class, subclass, objType int) {
	id := fmt.Sprintf("%2d/%d/%2d", class, subclass, objType)
	vm.cst.Set(id)
}

// RawData returns the text node for the raw byte string.
func (vm *ObjectsViewModel) RawData() *viewmodel.StringValueNode {
	return vm.rawData
}

// Properties returns the section for object properties.
func (vm *ObjectsViewModel) Properties() *viewmodel.SectionNode {
	return vm.properties
}

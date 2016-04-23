package console

// DetailDataChangeCallback is a callback function for the controller to
// notify changes in detail data.
type DetailDataChangeCallback func()

// NodeDetailViewFactory is a factory for creating detail views.
type NodeDetailViewFactory interface {
	// ForList creates a detail view for a single column list.
	ForList(controller ListDetailController, index int) DetailDataChangeCallback

	//forString(controller stringDetailController) detailDataChangeCallback
}

// NullDetailChangeCallback is the Null-Object implementation of DetailDataChangeCallback.
func NullDetailChangeCallback() {}

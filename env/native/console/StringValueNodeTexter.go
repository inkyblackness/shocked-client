package console

import (
	"fmt"

	"github.com/inkyblackness/shocked-client/viewmodel"
)

// StringValueNodeTexter is a texter for string values.
type StringValueNodeTexter struct {
	node *viewmodel.StringValueNode
}

// NewStringValueNodeTexter returns a new instance of StringValueNodeTexter.
func NewStringValueNodeTexter(node *viewmodel.StringValueNode, listener ViewModelListener) *StringValueNodeTexter {
	texter := &StringValueNodeTexter{node: node}

	node.Subscribe(func(string) {
		listener.OnMainDataChanged()
	})

	return texter
}

// Act implements the ViewModelNodeTexter interface.
func (texter *StringValueNodeTexter) Act(viewFactory NodeDetailViewFactory) {
}

// TextMain implements the ViewModelNodeTexter interface.
func (texter *StringValueNodeTexter) TextMain(addLine ViewModelLiner) {
	addLine(texter.node.Label(), fmt.Sprintf("  [%s]", texter.node.Get()), texter)
}

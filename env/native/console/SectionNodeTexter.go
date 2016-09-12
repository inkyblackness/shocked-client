package console

import (
	"github.com/inkyblackness/shocked-client/viewmodel"
)

// SectionNodeTexter is a texter for sections.
type SectionNodeTexter struct {
	node      *viewmodel.SectionNode
	listener  ViewModelListener
	subTexter []ViewModelNodeTexter
}

// NewSectionNodeTexter returns a new instance of SectionNodeTexter.
func NewSectionNodeTexter(node *viewmodel.SectionNode, listener ViewModelListener) *SectionNodeTexter {
	texter := &SectionNodeTexter{
		node:     node,
		listener: listener}

	texter.updateSubTexter()
	node.Subscribe(func() {
		texter.updateSubTexter()
		listener.OnMainDataChanged()
	})
	node.Available().Subscribe(func(bool) {
		listener.OnMainDataChanged()
	})

	return texter
}

// Act implements the ViewModelNodeTexter interface.
func (texter *SectionNodeTexter) Act(viewFactory NodeDetailViewFactory) {
}

// TextMain implements the ViewModelNodeTexter interface.
func (texter *SectionNodeTexter) TextMain(addLine ViewModelLiner) {
	if texter.node.Available().Get() {
		for _, subTexter := range texter.subTexter {
			subTexter.TextMain(addLine)
		}
	}
}

func (texter *SectionNodeTexter) updateSubTexter() {
	subNodes := texter.node.Get()
	texter.subTexter = make([]ViewModelNodeTexter, len(subNodes))
	for index, subNode := range subNodes {
		visitor := NewViewModelTexterVisitor(texter.listener)
		subNode.Specialize(visitor)
		texter.subTexter[index] = visitor.instance
	}
}

package modes

import (
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/ui"
)

type controlPanelBuilder struct {
	controlFactory controls.Factory
	parent         *ui.Area

	listLeft        ui.Anchor
	listRight       ui.Anchor
	listCenterEnd   ui.Anchor
	listCenterStart ui.Anchor

	lastBottom ui.Anchor
}

func newControlPanelBuilder(parent *ui.Area, controlFactory controls.Factory) *controlPanelBuilder {
	panelBuilder := &controlPanelBuilder{}
	panelBuilder.controlFactory = controlFactory
	panelBuilder.parent = parent
	panelBuilder.listLeft = ui.NewOffsetAnchor(parent.Left(), 2)
	panelBuilder.listRight = ui.NewOffsetAnchor(parent.Right(), -2)
	listCenter := ui.NewRelativeAnchor(panelBuilder.listLeft, panelBuilder.listRight, 0.5)
	panelBuilder.listCenterEnd = ui.NewOffsetAnchor(listCenter, -1)
	panelBuilder.listCenterStart = ui.NewOffsetAnchor(listCenter, 1)
	panelBuilder.lastBottom = ui.NewOffsetAnchor(parent.Top(), 0)

	return panelBuilder
}

func (panelBuilder *controlPanelBuilder) addComboProperty(labelText string, handler controls.SelectionChangeHandler) (label *controls.Label, box *controls.ComboBox) {
	top := ui.NewOffsetAnchor(panelBuilder.lastBottom, 2)
	bottom := ui.NewOffsetAnchor(top, 25)
	{
		builder := panelBuilder.controlFactory.ForLabel()
		builder.SetParent(panelBuilder.parent)
		builder.SetLeft(panelBuilder.listLeft)
		builder.SetTop(top)
		builder.SetRight(panelBuilder.listCenterEnd)
		builder.SetBottom(bottom)
		builder.AlignedHorizontallyBy(controls.RightAligner)
		label = builder.Build()
		label.SetText(labelText)
	}
	{
		builder := panelBuilder.controlFactory.ForComboBox()
		builder.SetParent(panelBuilder.parent)
		builder.SetLeft(panelBuilder.listCenterStart)
		builder.SetTop(top)
		builder.SetRight(panelBuilder.listRight)
		builder.SetBottom(bottom)
		builder.WithSelectionChangeHandler(handler)
		box = builder.Build()
	}
	panelBuilder.lastBottom = bottom

	return
}

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

func (panelBuilder *controlPanelBuilder) addTextureProperty(labelText string, provider controls.TextureProvider,
	handler controls.TextureSelectionChangeHandler) (label *controls.Label, selector *controls.TextureSelector) {
	top := ui.NewOffsetAnchor(panelBuilder.lastBottom, 2)
	bottom := ui.NewOffsetAnchor(top, 64+4)
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
		builder := panelBuilder.controlFactory.ForTextureSelector()
		builder.SetParent(panelBuilder.parent)
		builder.SetLeft(panelBuilder.listCenterStart)
		builder.SetTop(top)
		builder.SetRight(panelBuilder.listRight)
		builder.SetBottom(bottom)
		builder.WithProvider(provider)
		builder.WithSelectionChangeHandler(handler)
		selector = builder.Build()
	}
	panelBuilder.lastBottom = bottom

	return
}

func (panelBuilder *controlPanelBuilder) addSection(visible bool) (sectionArea *ui.Area, sectionBuilder *controlPanelBuilder) {
	sectionBuilder = &controlPanelBuilder{}
	sectionBuilder.controlFactory = panelBuilder.controlFactory
	sectionBuilder.listLeft = panelBuilder.listLeft
	sectionBuilder.listRight = panelBuilder.listRight
	sectionBuilder.listCenterEnd = panelBuilder.listCenterEnd
	sectionBuilder.listCenterStart = panelBuilder.listCenterStart
	sectionBuilder.lastBottom = panelBuilder.lastBottom
	{
		builder := ui.NewAreaBuilder()
		builder.SetParent(panelBuilder.parent)
		builder.SetLeft(ui.NewOffsetAnchor(panelBuilder.parent.Left(), 0))
		builder.SetTop(ui.NewOffsetAnchor(panelBuilder.lastBottom, 0))
		builder.SetRight(ui.NewOffsetAnchor(panelBuilder.parent.Right(), 0))
		builder.SetBottom(ui.NewResolvingAnchor(func() ui.Anchor {
			anchor := sectionBuilder.lastBottom
			if !sectionArea.IsVisible() {
				anchor = sectionArea.Top()
			}
			return anchor
		}))
		builder.SetVisible(visible)
		sectionArea = builder.Build()
		sectionBuilder.parent = sectionArea
	}
	panelBuilder.lastBottom = sectionArea.Bottom()

	return
}

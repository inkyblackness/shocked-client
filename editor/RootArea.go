package editor

import (
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/ui"
)

type rootArea struct {
	context Context
	area    *ui.Area

	modeBox      *controls.ComboBox
	messageLabel *controls.Label
}

func newRootArea(context Context) *ui.Area {
	root := &rootArea{context: context}
	areaBuilder := ui.NewAreaBuilder()

	areaBuilder.SetRight(ui.NewAbsoluteAnchor(0.0))
	areaBuilder.SetBottom(ui.NewAbsoluteAnchor(0.0))
	root.area = areaBuilder.Build()

	var topLine *ui.Area

	{
		builder := ui.NewAreaBuilder()
		top := ui.NewOffsetAnchor(root.area.Top(), 0)
		builder.SetParent(root.area)
		builder.SetLeft(ui.NewOffsetAnchor(root.area.Left(), 0))
		builder.SetTop(top)
		builder.SetRight(ui.NewOffsetAnchor(root.area.Right(), 0))
		builder.SetBottom(ui.NewOffsetAnchor(top, 25+4))
		topLine = builder.Build()
	}
	boxMessageSeparator := ui.NewOffsetAnchor(topLine.Left(), 250)
	{
		builder := context.ControlFactory().ForComboBox()
		builder.SetParent(topLine)
		builder.SetLeft(ui.NewOffsetAnchor(topLine.Left(), 2))
		builder.SetTop(ui.NewOffsetAnchor(topLine.Top(), 2))
		builder.SetRight(ui.NewOffsetAnchor(boxMessageSeparator, -2))
		builder.SetBottom(ui.NewOffsetAnchor(topLine.Bottom(), -2))
		builder.WithItems([]controls.ComboBoxItem{"Welcome (F1)", "Map Control (F1)", "Map Tiles (F2)", "Map Objects (F2)"})
		root.modeBox = builder.Build()
	}
	{
		builder := context.ControlFactory().ForLabel()
		builder.SetParent(topLine)
		builder.SetLeft(ui.NewOffsetAnchor(boxMessageSeparator, 2))
		builder.SetTop(ui.NewOffsetAnchor(topLine.Top(), 2))
		builder.SetRight(ui.NewOffsetAnchor(root.area.Right(), -2))
		builder.SetBottom(ui.NewOffsetAnchor(topLine.Bottom(), -2))
		builder.AlignedHorizontallyBy(controls.LeftAligner)
		root.messageLabel = builder.Build()
		context.ModelAdapter().OnMessageChanged(func() {
			root.messageLabel.SetText(context.ModelAdapter().Message())
		})
	}

	return root.area
}

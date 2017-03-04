package editor

import (
	"github.com/inkyblackness/shocked-client/editor/modes"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/ui"
)

type modeSelector struct {
	mode Mode
	name string
}

func (selector *modeSelector) String() string {
	return selector.name
}

type rootArea struct {
	context modes.Context
	area    *ui.Area

	modeArea *ui.Area

	modeBox      *controls.ComboBox
	messageLabel *controls.Label

	welcomeMode    *modeSelector
	mapControlMode *modeSelector
	allModes       []*modeSelector
}

func newRootArea(context modes.Context) *ui.Area {
	root := &rootArea{context: context}
	areaBuilder := ui.NewAreaBuilder()

	areaBuilder.SetRight(ui.NewAbsoluteAnchor(0.0))
	areaBuilder.SetBottom(ui.NewAbsoluteAnchor(0.0))
	root.area = areaBuilder.Build()

	var topLine *ui.Area

	topLineBottom := ui.NewOffsetAnchor(root.area.Top(), 25+4)
	{
		builder := ui.NewAreaBuilder()
		builder.SetParent(root.area)
		builder.SetLeft(ui.NewOffsetAnchor(root.area.Left(), 0))
		builder.SetTop(ui.NewOffsetAnchor(topLineBottom, 2))
		builder.SetRight(ui.NewOffsetAnchor(root.area.Right(), 0))
		builder.SetBottom(ui.NewOffsetAnchor(root.area.Bottom(), 0))
		root.modeArea = builder.Build()
	}
	{
		builder := ui.NewAreaBuilder()
		builder.SetParent(root.area)
		builder.SetLeft(ui.NewOffsetAnchor(root.area.Left(), 0))
		builder.SetTop(root.area.Top())
		builder.SetRight(ui.NewOffsetAnchor(root.area.Right(), 0))
		builder.SetBottom(topLineBottom)
		topLine = builder.Build()
	}

	root.welcomeMode = root.addMode(modes.NewWelcomeMode(context, root.modeArea), "Welcome")
	root.mapControlMode = root.addMode(modes.NewMapControlMode(context, root.modeArea), "Map Control")

	boxMessageSeparator := ui.NewOffsetAnchor(topLine.Left(), 250)
	{
		items := make([]controls.ComboBoxItem, len(root.allModes))
		for index, selector := range root.allModes {
			items[index] = selector
		}
		builder := context.ControlFactory().ForComboBox()
		builder.SetParent(topLine)
		builder.SetLeft(ui.NewOffsetAnchor(topLine.Left(), 2))
		builder.SetTop(ui.NewOffsetAnchor(topLine.Top(), 2))
		builder.SetRight(ui.NewOffsetAnchor(boxMessageSeparator, -2))
		builder.SetBottom(ui.NewOffsetAnchor(topLine.Bottom(), -2))
		builder.WithItems(items)
		builder.WithSelectionChangeHandler(func(item controls.ComboBoxItem) {
			root.setActiveMode(item.(*modeSelector))
		})
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

	root.setActiveMode(root.welcomeMode)

	return root.area
}

func (root *rootArea) addMode(mode Mode, name string) *modeSelector {
	selector := &modeSelector{
		mode: mode,
		name: name}

	root.allModes = append(root.allModes, selector)

	return selector
}

func (root *rootArea) setActiveMode(selector *modeSelector) {
	for _, other := range root.allModes {
		if other != selector {
			other.mode.SetActive(false)
		}
	}
	root.modeBox.SetSelectedItem(selector)
	selector.mode.SetActive(true)
}
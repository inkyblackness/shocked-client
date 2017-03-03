package modes

import (
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/ui"
)

// MapControlMode is a mode for archive map control.
type MapControlMode struct {
	context Context

	area *ui.Area

	activeLevelLabel *controls.Label
	activeLevelBox   *controls.ComboBox
}

// NewMapControlMode returns a new instance.
func NewMapControlMode(context Context, parent *ui.Area) *MapControlMode {
	mode := &MapControlMode{context: context}

	{
		builder := ui.NewAreaBuilder()
		builder.SetParent(parent)
		builder.SetLeft(ui.NewOffsetAnchor(parent.Left(), 0))
		builder.SetTop(ui.NewOffsetAnchor(parent.Top(), 0))
		builder.SetRight(ui.NewOffsetAnchor(parent.Left(), 300))
		builder.SetBottom(ui.NewOffsetAnchor(parent.Bottom(), 0))
		builder.SetVisible(false)
		builder.OnRender(func(area *ui.Area) {
			context.ForGraphics().RectangleRenderer().Fill(
				area.Left().Value(), area.Top().Value(), area.Right().Value(), area.Bottom().Value(),
				graphics.RGBA(0.7, 0.0, 0.7, 0.1))
		})
		mode.area = builder.Build()
	}
	listLeft := ui.NewOffsetAnchor(mode.area.Left(), 2)
	listRight := ui.NewOffsetAnchor(mode.area.Right(), 0)
	listCenter := ui.NewRelativeAnchor(listLeft, listRight, 0.5)
	listCenterEnd := ui.NewOffsetAnchor(listCenter, -1)
	listCenterStart := ui.NewOffsetAnchor(listCenter, 1)
	{
		top := ui.NewOffsetAnchor(mode.area.Top(), 2)
		{
			builder := context.ControlFactory().ForLabel()
			builder.SetParent(mode.area)
			builder.SetLeft(listLeft)
			builder.SetTop(top)
			builder.SetRight(listCenterEnd)
			builder.SetBottom(ui.NewOffsetAnchor(top, 25))
			builder.AlignedHorizontallyBy(controls.RightAligner)
			mode.activeLevelLabel = builder.Build()
			mode.activeLevelLabel.SetText("Active Level")
		}
		{
			builder := context.ControlFactory().ForComboBox()
			builder.SetParent(mode.area)
			builder.SetLeft(listCenterStart)
			builder.SetTop(top)
			builder.SetRight(listRight)
			builder.SetBottom(ui.NewOffsetAnchor(top, 25))
			builder.WithSelectionChangeHandler(func(item controls.ComboBoxItem) {
				context.ModelAdapter().RequestActiveLevel(item.(string))
			})
			mode.activeLevelBox = builder.Build()
		}
		adapter := context.ModelAdapter()
		activeLevelAdapter := adapter.ActiveLevel()
		activeLevelAdapter.OnIDChanged(func() {
			mode.activeLevelBox.SetSelectedItem(activeLevelAdapter.ID())
		})
		adapter.OnAvailableLevelsChanged(func() {
			ids := adapter.AvailableLevelIDs()
			items := make([]controls.ComboBoxItem, len(ids))
			for index, id := range ids {
				items[index] = id
			}
			mode.activeLevelBox.SetItems(items)
		})
	}

	return mode
}

// SetActive implements the Mode interface.
func (mode *MapControlMode) SetActive(active bool) {
	mode.area.SetVisible(active)
}

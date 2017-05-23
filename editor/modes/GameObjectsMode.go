package modes

import (
	"github.com/inkyblackness/shocked-client/editor/model"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/ui"
	"github.com/inkyblackness/shocked-client/ui/events"
)

// GameObjectsMode is a mode for archive level control.
type GameObjectsMode struct {
	context        Context
	objectsAdapter *model.ObjectsAdapter

	area *ui.Area

	selectedObjectLabel *controls.Label
	selectedObjectBox   *controls.ComboBox
	selectedObjectID    model.ObjectID
}

// NewGameObjectsMode returns a new instance.
func NewGameObjectsMode(context Context, parent *ui.Area) *GameObjectsMode {
	mode := &GameObjectsMode{
		context:        context,
		objectsAdapter: context.ModelAdapter().ObjectsAdapter(),

		selectedObjectID: model.ObjectID(0xFFFFFF)}

	{
		builder := ui.NewAreaBuilder()
		builder.SetParent(parent)
		builder.SetLeft(ui.NewOffsetAnchor(parent.Left(), 0))
		builder.SetTop(ui.NewOffsetAnchor(parent.Top(), 0))
		builder.SetRight(ui.NewRelativeAnchor(parent.Left(), parent.Right(), 0.66))
		builder.SetBottom(ui.NewOffsetAnchor(parent.Bottom(), 0))
		builder.SetVisible(false)
		builder.OnRender(func(area *ui.Area) {
			context.ForGraphics().RectangleRenderer().Fill(
				area.Left().Value(), area.Top().Value(), area.Right().Value(), area.Bottom().Value(),
				graphics.RGBA(0.7, 0.0, 0.7, 0.3))
		})
		builder.OnEvent(events.MouseMoveEventType, ui.SilentConsumer)
		builder.OnEvent(events.MouseButtonUpEventType, ui.SilentConsumer)
		builder.OnEvent(events.MouseButtonDownEventType, ui.SilentConsumer)
		builder.OnEvent(events.MouseButtonClickedEventType, ui.SilentConsumer)
		builder.OnEvent(events.MouseScrollEventType, ui.SilentConsumer)
		mode.area = builder.Build()
	}
	{
		panelBuilder := newControlPanelBuilder(mode.area, context.ControlFactory())

		{
			mode.selectedObjectLabel, mode.selectedObjectBox = panelBuilder.addComboProperty("Selected Object", func(item controls.ComboBoxItem) {
				typeItem := item.(*objectTypeItem)
				mode.onSelectedObjectTypeChanged(typeItem.id)
			})

			mode.objectsAdapter.OnObjectsChanged(mode.onObjectsChanged)
		}
	}

	return mode
}

// SetActive implements the Mode interface.
func (mode *GameObjectsMode) SetActive(active bool) {
	mode.area.SetVisible(active)
}

func (mode *GameObjectsMode) onObjectsChanged() {
	var items []controls.ComboBoxItem
	var selectedItem *objectTypeItem

	objects := mode.objectsAdapter.Objects()
	for _, object := range objects {
		newItem := &objectTypeItem{object.ID(), object.DisplayName()}
		items = append(items, newItem)
		if object.ID() == mode.selectedObjectID {
			selectedItem = newItem
		}
	}
	mode.selectedObjectBox.SetItems(items)
	if selectedItem != nil {
		mode.selectedObjectBox.SetSelectedItem(selectedItem)
		mode.onSelectedObjectTypeChanged(selectedItem.id)
	} else {
		mode.selectedObjectBox.SetSelectedItem(nil)
	}
}

func (mode *GameObjectsMode) onSelectedObjectTypeChanged(id model.ObjectID) {

}

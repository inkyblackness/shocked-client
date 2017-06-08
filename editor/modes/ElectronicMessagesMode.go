package modes

import (
	"github.com/inkyblackness/shocked-client/editor/model"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/ui"
	"github.com/inkyblackness/shocked-client/ui/events"

	dataModel "github.com/inkyblackness/shocked-model"
)

// ElectronicMessagesMode is a mode for game textures.
type ElectronicMessagesMode struct {
	context        Context
	messageAdapter *model.ElectronicMessageAdapter

	area           *ui.Area
	propertiesArea *ui.Area

	selectedMessageIDLabel  *controls.Label
	selectedMessageIDSlider *controls.Slider
	selectedMessageID       int

	propertiesHeader *controls.Label

	languageLabel    *controls.Label
	languageBox      *controls.ComboBox
	languageIndex    int
	verboseTextTitle *controls.Label
	verboseTextValue *controls.Label
}

// NewElectronicMessagesMode returns a new instance.
func NewElectronicMessagesMode(context Context, parent *ui.Area) *ElectronicMessagesMode {
	mode := &ElectronicMessagesMode{
		context:           context,
		messageAdapter:    context.ModelAdapter().ElectronicMessageAdapter(),
		selectedMessageID: -1}

	{
		builder := ui.NewAreaBuilder()
		builder.SetParent(parent)
		builder.SetLeft(ui.NewOffsetAnchor(parent.Left(), 0))
		builder.SetTop(ui.NewOffsetAnchor(parent.Top(), 0))
		builder.SetRight(ui.NewOffsetAnchor(parent.Right(), 0))
		builder.SetBottom(ui.NewOffsetAnchor(parent.Bottom(), 0))
		builder.SetVisible(false)
		mode.area = builder.Build()
	}
	{
		builder := ui.NewAreaBuilder()
		builder.SetParent(mode.area)
		builder.SetLeft(ui.NewOffsetAnchor(parent.Left(), 0))
		builder.SetTop(ui.NewOffsetAnchor(parent.Top(), 0))
		builder.SetRight(ui.NewRelativeAnchor(parent.Left(), parent.Right(), 0.5))
		builder.SetBottom(ui.NewRelativeAnchor(parent.Top(), parent.Bottom(), 0.5))
		builder.SetVisible(true)
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
		mode.propertiesArea = builder.Build()
	}
	{
		panelBuilder := newControlPanelBuilder(mode.propertiesArea, context.ControlFactory())

		{
			mode.selectedMessageIDLabel, mode.selectedMessageIDSlider = panelBuilder.addSliderProperty("Selected Texture ID",
				func(newValue int64) {
					mode.onMessageSelected(int(newValue))
				})
			mode.selectedMessageIDSlider.SetRange(0, 15)
		}
		mode.propertiesHeader = panelBuilder.addTitle("Properties")
		{
			mode.languageLabel, mode.languageBox = panelBuilder.addComboProperty("Language", mode.onLanguageChanged)
			items := []controls.ComboBoxItem{&enumItem{0, "STD"}, &enumItem{1, "FRA"}, &enumItem{2, "GER"}}
			mode.languageBox.SetItems(items)
			mode.languageBox.SetSelectedItem(items[0])

			mode.verboseTextTitle, mode.verboseTextValue = panelBuilder.addInfo("Verbose")
		}
	}
	mode.messageAdapter.OnMessageDataChanged(mode.onMessageDataChanged)

	return mode
}

// SetActive implements the Mode interface.
func (mode *ElectronicMessagesMode) SetActive(active bool) {
	mode.area.SetVisible(active)
}

func (mode *ElectronicMessagesMode) onMessageDataChanged() {
	mode.updateMessageText()
}

func (mode *ElectronicMessagesMode) onMessageSelected(id int) {
	mode.selectedMessageID = id
	mode.messageAdapter.RequestMessage(dataModel.ElectronicMessageTypeLog, id)
}

func (mode *ElectronicMessagesMode) onLanguageChanged(boxItem controls.ComboBoxItem) {
	item := boxItem.(*enumItem)
	mode.languageIndex = int(item.value)
	mode.updateMessageText()
}

func (mode *ElectronicMessagesMode) updateMessageText() {
	mode.verboseTextValue.SetText(mode.messageAdapter.VerboseText(mode.languageIndex))
}

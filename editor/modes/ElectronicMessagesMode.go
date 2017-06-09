package modes

import (
	"github.com/inkyblackness/shocked-client/editor/model"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/ui"
	"github.com/inkyblackness/shocked-client/ui/events"

	dataModel "github.com/inkyblackness/shocked-model"
)

// duplicate to ElectronicMessages.go - so far no need to transport this.
var messageRanges = map[dataModel.ElectronicMessageType]int64{
	dataModel.ElectronicMessageTypeMail:     0x09B8 - 0x0989,
	dataModel.ElectronicMessageTypeLog:      0x0A98 - 0x09B8,
	dataModel.ElectronicMessageTypeFragment: 0x0AA8 - 0x0A98}

// ElectronicMessagesMode is a mode for messages.
type ElectronicMessagesMode struct {
	context        Context
	messageAdapter *model.ElectronicMessageAdapter

	area *ui.Area

	propertiesArea *ui.Area

	messageTypeLabel        *controls.Label
	messageTypeBox          *controls.ComboBox
	messageType             dataModel.ElectronicMessageType
	messageTypeByIndex      map[uint32]dataModel.ElectronicMessageType
	selectedMessageIDLabel  *controls.Label
	selectedMessageIDSlider *controls.Slider
	selectedMessageID       int

	propertiesHeader *controls.Label

	languageLabel *controls.Label
	languageBox   *controls.ComboBox
	languageIndex int
	variantLabel  *controls.Label
	variantBox    *controls.ComboBox
	variantTerse  bool

	displayArea *ui.Area

	textValue *controls.Label
}

// NewElectronicMessagesMode returns a new instance.
func NewElectronicMessagesMode(context Context, parent *ui.Area) *ElectronicMessagesMode {
	mode := &ElectronicMessagesMode{
		context:            context,
		messageAdapter:     context.ModelAdapter().ElectronicMessageAdapter(),
		messageTypeByIndex: make(map[uint32]dataModel.ElectronicMessageType),
		selectedMessageID:  -1}

	indexByMessageType := make(map[dataModel.ElectronicMessageType]uint32)
	for index, messageType := range dataModel.ElectronicMessageTypes() {
		mode.messageTypeByIndex[uint32(index)] = messageType
		indexByMessageType[messageType] = uint32(index)
	}

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
			mode.messageTypeLabel, mode.messageTypeBox = panelBuilder.addComboProperty("Message Type", mode.onMessageTypeChanged)
			items := []controls.ComboBoxItem{
				&enumItem{indexByMessageType[dataModel.ElectronicMessageTypeMail], "Mail"},
				&enumItem{indexByMessageType[dataModel.ElectronicMessageTypeLog], "Log"},
				&enumItem{indexByMessageType[dataModel.ElectronicMessageTypeFragment], "Fragment"}}
			mode.messageTypeBox.SetItems(items)
		}
		{
			mode.selectedMessageIDLabel, mode.selectedMessageIDSlider = panelBuilder.addSliderProperty("Selected Message ID",
				func(newValue int64) { mode.onMessageSelected(int(newValue)) })
		}
		mode.propertiesHeader = panelBuilder.addTitle("Properties")
		{
			mode.languageLabel, mode.languageBox = panelBuilder.addComboProperty("Language", mode.onLanguageChanged)
			items := []controls.ComboBoxItem{&enumItem{0, "STD"}, &enumItem{1, "FRA"}, &enumItem{2, "GER"}}
			mode.languageBox.SetItems(items)
			mode.languageBox.SetSelectedItem(items[0])
		}
		{
			mode.variantLabel, mode.variantBox = panelBuilder.addComboProperty("Text Variant", mode.onVariantChanged)
			items := []controls.ComboBoxItem{&enumItem{0, "Verbose"}, &enumItem{1, "Terse"}}
			mode.variantBox.SetItems(items)
			mode.variantBox.SetSelectedItem(items[0])
		}
	}
	{
		builder := ui.NewAreaBuilder()
		builder.SetParent(mode.area)
		builder.SetLeft(ui.NewOffsetAnchor(parent.Left(), 0))
		builder.SetTop(ui.NewRelativeAnchor(parent.Top(), parent.Bottom(), 0.5))
		builder.SetRight(ui.NewOffsetAnchor(parent.Right(), 0))
		builder.SetBottom(ui.NewOffsetAnchor(parent.Bottom(), 0))
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
		mode.displayArea = builder.Build()
	}
	{
		labelBuilder := mode.context.ControlFactory().ForLabel()

		labelBuilder.SetParent(mode.displayArea)
		labelBuilder.SetTop(ui.NewOffsetAnchor(mode.displayArea.Top(), 5))
		labelBuilder.SetBottom(ui.NewOffsetAnchor(mode.displayArea.Bottom(), -5))
		labelBuilder.SetLeft(ui.NewRelativeAnchor(mode.displayArea.Left(), mode.displayArea.Right(), 0.25))
		labelBuilder.SetRight(ui.NewRelativeAnchor(mode.displayArea.Left(), mode.displayArea.Right(), 0.75))
		labelBuilder.AlignedHorizontallyBy(controls.LeftAligner)
		labelBuilder.AlignedVerticallyBy(controls.LeftAligner)
		labelBuilder.SetFitToWidth()
		mode.textValue = labelBuilder.Build()
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

func (mode *ElectronicMessagesMode) onMessageTypeChanged(boxItem controls.ComboBoxItem) {
	item := boxItem.(*enumItem)
	mode.messageType = mode.messageTypeByIndex[item.value]
	mode.selectedMessageIDSlider.SetValue(0)
	mode.selectedMessageIDSlider.SetRange(0, messageRanges[mode.messageType]-1)
	mode.onMessageSelected(0)
}

func (mode *ElectronicMessagesMode) onMessageSelected(id int) {
	mode.selectedMessageID = id
	mode.messageAdapter.RequestMessage(mode.messageType, id)
}

func (mode *ElectronicMessagesMode) onLanguageChanged(boxItem controls.ComboBoxItem) {
	item := boxItem.(*enumItem)
	mode.languageIndex = int(item.value)
	mode.updateMessageText()
}

func (mode *ElectronicMessagesMode) onVariantChanged(boxItem controls.ComboBoxItem) {
	item := boxItem.(*enumItem)
	mode.variantTerse = item.value != 0
	mode.updateMessageText()
}

func (mode *ElectronicMessagesMode) updateMessageText() {
	text := ""
	if mode.variantTerse {
		text = mode.messageAdapter.TerseText(mode.languageIndex)
	} else {
		text = mode.messageAdapter.VerboseText(mode.languageIndex)
	}
	mode.textValue.SetText(text)
}
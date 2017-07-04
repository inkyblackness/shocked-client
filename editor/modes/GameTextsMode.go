package modes

import (
	"github.com/inkyblackness/shocked-client/editor/model"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/ui"
	"github.com/inkyblackness/shocked-client/ui/events"

	dataModel "github.com/inkyblackness/shocked-model"
)

// GameTextsMode is a mode for arbitrary game texts.
type GameTextsMode struct {
	context     Context
	textAdapter *model.TextAdapter

	area           *ui.Area
	propertiesArea *ui.Area

	typeLabel            *controls.Label
	typeBox              *controls.ComboBox
	selectedResourceType dataModel.ResourceType

	languageLabel *controls.Label
	languageBox   *controls.ComboBox
	language      dataModel.ResourceLanguage

	selectedTextIDLabel  *controls.Label
	selectedTextIDSlider *controls.Slider
	selectedTextID       int

	textDrop  *ui.Area
	textValue *controls.Label
}

// NewGameTextsMode returns a new instance.
func NewGameTextsMode(context Context, parent *ui.Area) *GameTextsMode {
	mode := &GameTextsMode{
		context:        context,
		textAdapter:    context.ModelAdapter().TextAdapter(),
		selectedTextID: -1}

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
		mode.propertiesArea = builder.Build()
	}
	{
		panelBuilder := newControlPanelBuilder(mode.propertiesArea, context.ControlFactory())
		var initialTypeItem controls.ComboBoxItem
		var initialLanguageItem controls.ComboBoxItem

		{
			mode.typeLabel, mode.typeBox = panelBuilder.addComboProperty("Text Type", mode.onTextTypeChanged)
			items := []controls.ComboBoxItem{
				&enumItem{uint32(dataModel.ResourceTypeTrapMessages), "Trap Messages"},
				&enumItem{uint32(dataModel.ResourceTypeWords), "Words"},
				&enumItem{uint32(dataModel.ResourceTypeLogCategories), "Log Categories"},
				&enumItem{uint32(dataModel.ResourceTypeScreenMessages), "Screen Messages"},
				&enumItem{uint32(dataModel.ResourceTypeInfoNodeMessages), "Info Node Messages (8/5/6)"},
				&enumItem{uint32(dataModel.ResourceTypeAccessCardNames), "Access Card Names"},
				&enumItem{uint32(dataModel.ResourceTypeDataletMessages), "Datalet Messages (8/5/8)"},
				&enumItem{uint32(dataModel.ResourceTypePaperTexts), "Paper Texts"}}

			mode.typeBox.SetItems(items)
			initialTypeItem = items[0]
		}
		{
			mode.languageLabel, mode.languageBox = panelBuilder.addComboProperty("Language", mode.onLanguageChanged)
			items := []controls.ComboBoxItem{
				&enumItem{uint32(dataModel.ResourceLanguageStandard), "STD"},
				&enumItem{uint32(dataModel.ResourceLanguageFrench), "FRN"},
				&enumItem{uint32(dataModel.ResourceLanguageGerman), "GER"}}
			mode.languageBox.SetItems(items)
			initialLanguageItem = items[0]
		}
		{
			mode.selectedTextIDLabel, mode.selectedTextIDSlider = panelBuilder.addSliderProperty("Selected Text ID",
				func(newValue int64) {
					mode.onTextSelected(int(newValue))
				})
		}
		mode.languageBox.SetSelectedItem(initialLanguageItem)
		mode.onLanguageChanged(initialLanguageItem)
		mode.typeBox.SetSelectedItem(initialTypeItem)
		mode.onTextTypeChanged(initialTypeItem)
	}
	{
		padding := float32(5.0)

		{
			dropBuilder := ui.NewAreaBuilder()
			displayBuilder := mode.context.ControlFactory().ForLabel()
			left := ui.NewOffsetAnchor(mode.propertiesArea.Right(), padding)
			right := ui.NewOffsetAnchor(mode.area.Right(), -padding)
			top := ui.NewOffsetAnchor(mode.area.Top(), padding)
			bottom := ui.NewOffsetAnchor(mode.area.Bottom(), -padding)

			dropBuilder.SetParent(mode.area)
			displayBuilder.SetParent(mode.area)
			dropBuilder.SetLeft(left)
			displayBuilder.SetLeft(left)
			dropBuilder.SetRight(right)
			displayBuilder.SetRight(right)
			dropBuilder.SetTop(top)
			displayBuilder.SetTop(top)
			dropBuilder.SetBottom(bottom)
			displayBuilder.SetBottom(bottom)
			displayBuilder.AlignedHorizontallyBy(controls.LeftAligner)
			displayBuilder.AlignedVerticallyBy(controls.LeftAligner)
			displayBuilder.SetFitToWidth()
			mode.textDrop = dropBuilder.Build()
			mode.textValue = displayBuilder.Build()
			mode.textValue.AllowTextChange(mode.onTextModified)
		}
	}
	mode.context.ModelAdapter().OnProjectChanged(func() {
		mode.requestText()
	})
	mode.textAdapter.OnTextChanged(mode.onTextChanged)

	return mode
}

// SetActive implements the Mode interface.
func (mode *GameTextsMode) SetActive(active bool) {
	mode.area.SetVisible(active)
}

func (mode *GameTextsMode) onTextTypeChanged(boxItem controls.ComboBoxItem) {
	item := boxItem.(*enumItem)
	mode.selectedResourceType = dataModel.ResourceType(item.value)

	mode.onTextSelected(0)
	mode.selectedTextIDSlider.SetRange(0, int64(dataModel.MaxEntriesFor(mode.selectedResourceType))-1)
	mode.selectedTextIDSlider.SetValue(0)
	mode.requestText()
}

func (mode *GameTextsMode) onLanguageChanged(boxItem controls.ComboBoxItem) {
	item := boxItem.(*enumItem)
	mode.language = dataModel.ResourceLanguage(item.value)
	mode.requestText()
}

func (mode *GameTextsMode) onTextSelected(id int) {
	mode.selectedTextID = id
	mode.requestText()
}

func (mode *GameTextsMode) requestText() {
	key := dataModel.MakeLocalizedResourceKey(mode.selectedResourceType, mode.language, uint16(mode.selectedTextID))
	mode.textAdapter.RequestText(key)
}

func (mode *GameTextsMode) onTextChanged() {
	mode.textValue.SetText(mode.textAdapter.Text())
}

func (mode *GameTextsMode) onTextModified(newText string) {
	mode.textAdapter.RequestTextChange(newText)
}

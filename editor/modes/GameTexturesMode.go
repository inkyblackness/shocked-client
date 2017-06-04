package modes

import (
	"github.com/inkyblackness/shocked-client/editor/model"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/ui"
	"github.com/inkyblackness/shocked-client/ui/events"

	dataModel "github.com/inkyblackness/shocked-model"
)

// GameTexturesMode is a mode for game textures.
type GameTexturesMode struct {
	context        Context
	textureAdapter *model.TextureAdapter

	area *ui.Area

	selectedTextureLabel    *controls.Label
	selectedTextureSelector *controls.TextureSelector
	selectedTextureIDLabel  *controls.Label
	selectedTextureIDSlider *controls.Slider
	selectedTextureID       int

	climbableLabel *controls.Label
	climbableBox   *controls.ComboBox
	climbableItems []controls.ComboBoxItem
}

// NewGameTexturesMode returns a new instance.
func NewGameTexturesMode(context Context, parent *ui.Area) *GameTexturesMode {
	mode := &GameTexturesMode{
		context:           context,
		textureAdapter:    context.ModelAdapter().TextureAdapter(),
		selectedTextureID: -1}

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
			mode.selectedTextureIDLabel, mode.selectedTextureIDSlider = panelBuilder.addSliderProperty("Selected Texture ID",
				func(newValue int64) {
					mode.selectedTextureSelector.SetSelectedIndex(int(newValue))
					mode.onTextureSelected(int(newValue))
				})
			mode.selectedTextureLabel, mode.selectedTextureSelector = panelBuilder.addTextureProperty("Selected Texture",
				mode.worldTextures, func(newValue int) {
					mode.selectedTextureIDSlider.SetValue(int64(newValue))
					mode.onTextureSelected(newValue)
				})
		}
		{
			mode.climbableLabel, mode.climbableBox = panelBuilder.addComboProperty("Climbable", mode.onClimbableChanged)
			mode.climbableItems = []controls.ComboBoxItem{&enumItem{0, "No"}, &enumItem{1, "Yes"}}
			mode.climbableBox.SetItems(mode.climbableItems)
		}
	}
	mode.textureAdapter.OnGameTexturesChanged(mode.onGameTexturesChanged)

	return mode
}

// SetActive implements the Mode interface.
func (mode *GameTexturesMode) SetActive(active bool) {
	mode.area.SetVisible(active)
}

func (mode *GameTexturesMode) worldTextures() []*graphics.BitmapTexture {
	textureCount := mode.context.ModelAdapter().TextureAdapter().WorldTextureCount()
	textures := make([]*graphics.BitmapTexture, textureCount)
	store := mode.context.ForGraphics().WorldTextureStore(dataModel.TextureLarge)

	for index := 0; index < textureCount; index++ {
		textures[index] = store.Texture(graphics.TextureKeyFromInt(index))
	}

	return textures
}

func (mode *GameTexturesMode) onGameTexturesChanged() {
	textureCount := mode.textureAdapter.WorldTextureCount()

	mode.selectedTextureIDSlider.SetRange(0, int64(textureCount)-1)
}

func (mode *GameTexturesMode) onTextureSelected(id int) {
	texture := mode.textureAdapter.GameTexture(id)

	mode.selectedTextureID = id
	if texture.Climbable() {
		mode.climbableBox.SetSelectedItem(mode.climbableItems[1])
	} else {
		mode.climbableBox.SetSelectedItem(mode.climbableItems[0])
	}
}

func (mode *GameTexturesMode) requestTexturePropertiesChange(modifier func(*dataModel.TextureProperties)) {
	var properties dataModel.TextureProperties

	modifier(&properties)
	mode.textureAdapter.RequestTexturePropertiesChange(mode.selectedTextureID, &properties)
}

func (mode *GameTexturesMode) onClimbableChanged(boxItem controls.ComboBoxItem) {
	item := boxItem.(*enumItem)
	mode.requestTexturePropertiesChange(func(properties *dataModel.TextureProperties) {
		properties.Climbable = boolAsPointer(item.value != 0)
	})
}

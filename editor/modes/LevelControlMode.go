package modes

import (
	"fmt"

	"github.com/inkyblackness/shocked-client/editor/display"
	"github.com/inkyblackness/shocked-client/editor/model"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/ui"
	"github.com/inkyblackness/shocked-client/ui/events"

	dataModel "github.com/inkyblackness/shocked-model"
)

// LevelControlMode is a mode for archive level control.
type LevelControlMode struct {
	context      Context
	levelAdapter *model.LevelAdapter

	mapDisplay *display.MapDisplay

	area *ui.Area

	activeLevelLabel *controls.Label
	activeLevelBox   *controls.ComboBox

	heightShiftLabel *controls.Label
	heightShiftBox   *controls.ComboBox

	levelTexturesLabel       *controls.Label
	levelTexturesSelector    *controls.TextureSelector
	currentLevelTextureIndex int
	worldTexturesLabel       *controls.Label
	worldTexturesSelector    *controls.TextureSelector

	selectedSurveillanceIndex    int
	surveillanceIndexLabel       *controls.Label
	surveillanceIndexBox         *controls.ComboBox
	surveillanceSourceLabel      *controls.Label
	surveillanceSourceSlider     *controls.Slider
	surveillanceDeathwatchLabel  *controls.Label
	surveillanceDeathwatchSlider *controls.Slider
}

// NewLevelControlMode returns a new instance.
func NewLevelControlMode(context Context, parent *ui.Area, mapDisplay *display.MapDisplay) *LevelControlMode {
	mode := &LevelControlMode{
		context:                   context,
		levelAdapter:              context.ModelAdapter().ActiveLevel(),
		mapDisplay:                mapDisplay,
		currentLevelTextureIndex:  -1,
		selectedSurveillanceIndex: -1}

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
			mode.activeLevelLabel, mode.activeLevelBox = panelBuilder.addComboProperty("Active Level", func(item controls.ComboBoxItem) {
				context.ModelAdapter().RequestActiveLevel(item.(int))
			})

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
		{
			mode.heightShiftLabel, mode.heightShiftBox = panelBuilder.addComboProperty("Tile Height", mode.onHeightShiftChanged)
			heightShiftItems := make([]controls.ComboBoxItem, 8)

			heightShiftItems[0] = &enumItem{0, "32 Tiles"}
			heightShiftItems[1] = &enumItem{1, "16 Tiles"}
			heightShiftItems[2] = &enumItem{2, "8 Tiles"}
			heightShiftItems[3] = &enumItem{3, "4 Tiles"}
			heightShiftItems[4] = &enumItem{4, "2 Tiles"}
			heightShiftItems[5] = &enumItem{5, "1 Tile"}
			heightShiftItems[6] = &enumItem{6, "1/2 Tile"}
			heightShiftItems[7] = &enumItem{7, "1/4 Tile"}
			mode.heightShiftBox.SetItems(heightShiftItems)
			mode.levelAdapter.OnLevelPropertiesChanged(func() {
				heightShift := mode.levelAdapter.HeightShift()
				if (heightShift >= 0) && (heightShift < len(heightShiftItems)) {
					mode.heightShiftBox.SetSelectedItem(heightShiftItems[heightShift])
				} else {
					mode.heightShiftBox.SetSelectedItem(nil)
				}
			})
		}
		{
			mode.levelTexturesLabel, mode.levelTexturesSelector = panelBuilder.addTextureProperty("Level Textures",
				mode.levelTextures, mode.onSelectedLevelTextureChanged)
			mode.worldTexturesLabel, mode.worldTexturesSelector = panelBuilder.addTextureProperty("World Textures",
				mode.worldTextures, mode.onSelectedWorldTextureChanged)
		}
		{
			mode.surveillanceIndexLabel, mode.surveillanceIndexBox =
				panelBuilder.addComboProperty("Surveillance Object", mode.onSurveillanceIndexChanged)
			mode.surveillanceSourceLabel, mode.surveillanceSourceSlider =
				panelBuilder.addSliderProperty("Surveillance Source", mode.onSurveillanceSourceChanged)
			mode.surveillanceDeathwatchLabel, mode.surveillanceDeathwatchSlider =
				panelBuilder.addSliderProperty("Surveillance Deathwatch", mode.onSurveillanceDeathwatchChanged)
			mode.surveillanceSourceSlider.SetRange(0, 871)
			mode.surveillanceDeathwatchSlider.SetRange(0, 871)

			mode.levelAdapter.OnLevelSurveillanceChanged(mode.onLevelSurveillanceChanged)
		}
	}

	return mode
}

// SetActive implements the Mode interface.
func (mode *LevelControlMode) SetActive(active bool) {
	mode.area.SetVisible(active)
	mode.mapDisplay.SetVisible(active)
}

func (mode *LevelControlMode) levelTextures() []*graphics.BitmapTexture {
	ids := mode.context.ModelAdapter().ActiveLevel().LevelTextureIDs()
	textures := make([]*graphics.BitmapTexture, len(ids))
	store := mode.context.ForGraphics().WorldTextureStore(dataModel.TextureLarge)

	for index, id := range ids {
		textures[index] = store.Texture(graphics.TextureKeyFromInt(id))
	}

	return textures
}

func (mode *LevelControlMode) worldTextures() []*graphics.BitmapTexture {
	textureCount := mode.context.ModelAdapter().TextureAdapter().WorldTextureCount()
	textures := make([]*graphics.BitmapTexture, textureCount)
	store := mode.context.ForGraphics().WorldTextureStore(dataModel.TextureLarge)

	for index := 0; index < textureCount; index++ {
		textures[index] = store.Texture(graphics.TextureKeyFromInt(index))
	}

	return textures
}

func (mode *LevelControlMode) onHeightShiftChanged(boxItem controls.ComboBoxItem) {
	item := boxItem.(*enumItem)
	mode.levelAdapter.RequestLevelPropertiesChange(func(properties *dataModel.LevelProperties) {
		properties.HeightShift = intAsPointer(int(item.value))
	})
}

func (mode *LevelControlMode) onSelectedLevelTextureChanged(index int) {
	ids := mode.context.ModelAdapter().ActiveLevel().LevelTextureIDs()

	if (index >= 0) && (index < len(ids)) {
		mode.worldTexturesSelector.SetSelectedIndex(ids[index])
		mode.currentLevelTextureIndex = index
	} else {
		mode.worldTexturesSelector.SetSelectedIndex(-1)
		mode.currentLevelTextureIndex = -1
	}
}

func (mode *LevelControlMode) onSelectedWorldTextureChanged(index int) {
	levelAdapter := mode.context.ModelAdapter().ActiveLevel()
	ids := levelAdapter.LevelTextureIDs()

	if (mode.currentLevelTextureIndex >= 0) && (mode.currentLevelTextureIndex < len(ids)) {
		newIDs := make([]int, len(ids))
		copy(newIDs, ids)
		newIDs[mode.currentLevelTextureIndex] = index
		levelAdapter.RequestLevelTexturesChange(newIDs)
	}
}

func (mode *LevelControlMode) onLevelSurveillanceChanged() {
	surveillanceCount := mode.levelAdapter.ObjectSurveillanceCount()
	items := make([]controls.ComboBoxItem, surveillanceCount)
	var selectedItem controls.ComboBoxItem

	for index := 0; index < surveillanceCount; index++ {
		item := &enumItem{uint32(index), fmt.Sprintf("Object %v", index)}
		items[index] = item
		if index == mode.selectedSurveillanceIndex {
			selectedItem = item
		}
	}

	mode.surveillanceIndexBox.SetItems(items)
	mode.surveillanceIndexBox.SetSelectedItem(selectedItem)
	mode.onSurveillanceIndexChanged(selectedItem)
}

func (mode *LevelControlMode) onSurveillanceIndexChanged(boxItem controls.ComboBoxItem) {
	if boxItem != nil {
		item := boxItem.(*enumItem)
		mode.selectedSurveillanceIndex = int(item.value)
		sourceIndex, deathwatchIndex := mode.levelAdapter.ObjectSurveillanceInfo(mode.selectedSurveillanceIndex)
		mode.surveillanceSourceSlider.SetValue(int64(sourceIndex))
		mode.surveillanceDeathwatchSlider.SetValue(int64(deathwatchIndex))
	} else {
		mode.surveillanceSourceSlider.SetValueUndefined()
		mode.surveillanceDeathwatchSlider.SetValueUndefined()
		mode.selectedSurveillanceIndex = -1
	}
}

func (mode *LevelControlMode) onSurveillanceSourceChanged(newValue int64) {
	newIndex := int(newValue)
	mode.levelAdapter.RequestObjectSurveillance(mode.selectedSurveillanceIndex, &newIndex, nil)
}

func (mode *LevelControlMode) onSurveillanceDeathwatchChanged(newValue int64) {
	newIndex := int(newValue)
	mode.levelAdapter.RequestObjectSurveillance(mode.selectedSurveillanceIndex, nil, &newIndex)
}

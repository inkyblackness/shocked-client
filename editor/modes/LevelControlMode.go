package modes

import (
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/ui"

	dataModel "github.com/inkyblackness/shocked-model"
)

// LevelControlMode is a mode for archive level control.
type LevelControlMode struct {
	context Context

	area *ui.Area

	activeLevelLabel *controls.Label
	activeLevelBox   *controls.ComboBox

	levelTexturesLabel    *controls.Label
	levelTexturesSelector *controls.TextureSelector
	worldTexturesLabel    *controls.Label
	worldTexturesSelector *controls.TextureSelector
}

// NewLevelControlMode returns a new instance.
func NewLevelControlMode(context Context, parent *ui.Area) *LevelControlMode {
	mode := &LevelControlMode{context: context}

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
				graphics.RGBA(0.7, 0.0, 0.7, 0.1))
		})
		mode.area = builder.Build()
	}
	{
		panelBuilder := newControlPanelBuilder(mode.area, context.ControlFactory())

		{
			mode.activeLevelLabel, mode.activeLevelBox = panelBuilder.addComboProperty("Active Level", func(item controls.ComboBoxItem) {
				context.ModelAdapter().RequestActiveLevel(item.(string))
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
			mode.levelTexturesLabel, mode.levelTexturesSelector = panelBuilder.addTextureProperty("Level Textures",
				mode.levelTextures, mode.onSelectedLevelTextureChanged)
			mode.worldTexturesLabel, mode.worldTexturesSelector = panelBuilder.addTextureProperty("World Textures",
				mode.worldTextures, mode.onSelectedWorldTextureChanged)
		}
	}

	return mode
}

// SetActive implements the Mode interface.
func (mode *LevelControlMode) SetActive(active bool) {
	mode.area.SetVisible(active)
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

func (mode *LevelControlMode) onSelectedLevelTextureChanged(index int) {
	ids := mode.context.ModelAdapter().ActiveLevel().LevelTextureIDs()

	if (index >= 0) && (index < len(ids)) {
		mode.worldTexturesSelector.SetSelectedIndex(ids[index])
	} else {
		mode.worldTexturesSelector.SetSelectedIndex(-1)
	}
}

func (mode *LevelControlMode) onSelectedWorldTextureChanged(index int) {

}

package modes

import (
	"fmt"

	"github.com/inkyblackness/shocked-client/editor/display"
	"github.com/inkyblackness/shocked-client/editor/model"
	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/env/keys"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/ui"
	"github.com/inkyblackness/shocked-client/ui/events"
	"github.com/inkyblackness/shocked-client/util"
	dataModel "github.com/inkyblackness/shocked-model"
)

type tilePropertySetter func(properties *dataModel.TileProperties, value interface{})
type tilePropertyItem struct {
	value  interface{}
	setter tilePropertySetter
}

func (item *tilePropertyItem) String() string {
	return fmt.Sprintf("%v", item.value)
}

// LevelMapMode is a mode for level maps.
type LevelMapMode struct {
	context Context

	mapDisplay *display.MapDisplay

	area       *ui.Area
	panel      *ui.Area
	panelRight ui.Anchor

	selectedTiles []model.TileCoordinate

	tileTypeLabel      *controls.Label
	tileTypeBox        *controls.ComboBox
	tileTypeItems      map[dataModel.TileType]*tilePropertyItem
	floorHeightLabel   *controls.Label
	floorHeightBox     *controls.ComboBox
	ceilingHeightLabel *controls.Label
	ceilingHeightBox   *controls.ComboBox
	slopeHeightLabel   *controls.Label
	slopeHeightBox     *controls.ComboBox
}

// NewLevelMapMode returns a new instance.
func NewLevelMapMode(context Context, parent *ui.Area, mapDisplay *display.MapDisplay) *LevelMapMode {
	mode := &LevelMapMode{
		context:    context,
		mapDisplay: mapDisplay}

	{
		builder := ui.NewAreaBuilder()
		builder.SetParent(parent)
		builder.SetLeft(ui.NewOffsetAnchor(parent.Left(), 0))
		builder.SetTop(ui.NewOffsetAnchor(parent.Top(), 0))
		builder.SetRight(ui.NewOffsetAnchor(parent.Right(), 0))
		builder.SetBottom(ui.NewOffsetAnchor(parent.Bottom(), 0))
		builder.SetVisible(false)
		builder.OnEvent(events.MouseMoveEventType, mode.onMouseMoved)
		builder.OnEvent(events.MouseButtonClickedEventType, mode.onMouseButtonClicked)
		mode.area = builder.Build()
	}
	{
		minRight := ui.NewOffsetAnchor(mode.area.Left(), 100)
		maxRight := ui.NewRelativeAnchor(mode.area.Left(), mode.area.Right(), 0.5)
		mode.panelRight = ui.NewLimitedAnchor(minRight, maxRight, ui.NewOffsetAnchor(mode.area.Left(), 200))
		builder := ui.NewAreaBuilder()
		builder.SetParent(mode.area)
		builder.SetLeft(ui.NewOffsetAnchor(mode.area.Left(), 0))
		builder.SetTop(ui.NewOffsetAnchor(mode.area.Top(), 0))
		builder.SetRight(mode.panelRight)
		builder.SetBottom(ui.NewOffsetAnchor(mode.area.Bottom(), 0))
		builder.OnRender(func(area *ui.Area) {
			context.ForGraphics().RectangleRenderer().Fill(
				area.Left().Value(), area.Top().Value(), area.Right().Value(), area.Bottom().Value(),
				graphics.RGBA(0.7, 0.0, 0.7, 0.3))
		})

		builder.OnEvent(events.MouseButtonClickedEventType, ui.SilentConsumer)
		builder.OnEvent(events.MouseScrollEventType, ui.SilentConsumer)

		lastGrabX := float32(0.0)
		builder.OnEvent(events.MouseButtonDownEventType, func(area *ui.Area, event events.Event) bool {
			buttonEvent := event.(*events.MouseButtonEvent)
			if buttonEvent.Buttons() == env.MousePrimary {
				area.RequestFocus()
				lastGrabX, _ = buttonEvent.Position()
			}
			return true
		})
		builder.OnEvent(events.MouseButtonUpEventType, func(area *ui.Area, event events.Event) bool {
			buttonEvent := event.(*events.MouseButtonEvent)
			if buttonEvent.AffectedButtons() == env.MousePrimary {
				area.ReleaseFocus()
			}
			return true
		})
		builder.OnEvent(events.MouseMoveEventType, func(area *ui.Area, event events.Event) bool {
			moveEvent := event.(*events.MouseMoveEvent)
			if area.HasFocus() {
				newX, _ := moveEvent.Position()
				mode.panelRight.RequestValue(mode.panelRight.Value() + (newX - lastGrabX))
				lastGrabX = newX
			}
			return true
		})

		mode.panel = builder.Build()
	}

	{
		panelBuilder := newControlPanelBuilder(mode.panel, context.ControlFactory())

		mode.tileTypeLabel, mode.tileTypeBox = panelBuilder.addComboProperty("Tile Type", mode.onTilePropertyChangeRequested)
		{
			setter := func(properties *dataModel.TileProperties, value interface{}) {
				tileType := value.(dataModel.TileType)
				properties.Type = &tileType
			}
			tileTypes := dataModel.TileTypes()
			tileTypeItems := make([]controls.ComboBoxItem, len(tileTypes))
			mode.tileTypeItems = make(map[dataModel.TileType]*tilePropertyItem)
			for index, tileType := range tileTypes {
				item := &tilePropertyItem{tileType, setter}
				tileTypeItems[index] = item
				mode.tileTypeItems[tileType] = item
			}
			mode.tileTypeItems[dataModel.TileType("")] = &tilePropertyItem{"", nil}
			mode.tileTypeBox.SetItems(tileTypeItems)
		}

		mode.floorHeightLabel, mode.floorHeightBox = panelBuilder.addComboProperty("Floor Height", mode.onTilePropertyChangeRequested)
		mode.ceilingHeightLabel, mode.ceilingHeightBox = panelBuilder.addComboProperty("Ceiling Height", mode.onTilePropertyChangeRequested)
		mode.slopeHeightLabel, mode.slopeHeightBox = panelBuilder.addComboProperty("Slope Height", mode.onTilePropertyChangeRequested)
	}

	return mode
}

// SetActive implements the Mode interface.
func (mode *LevelMapMode) SetActive(active bool) {
	if active {
		mode.mapDisplay.SetSelectedTiles(mode.selectedTiles)
	} else {
		mode.mapDisplay.ClearHighlightedTile()
		mode.mapDisplay.SetSelectedTiles(nil)
	}
	mode.area.SetVisible(active)
	mode.mapDisplay.SetVisible(active)
}

func (mode *LevelMapMode) onMouseMoved(area *ui.Area, event events.Event) (consumed bool) {
	mouseEvent := event.(*events.MouseMoveEvent)

	if mouseEvent.Buttons() == 0 {
		worldX, worldY := mode.mapDisplay.WorldCoordinatesForPixel(mouseEvent.Position())
		coord := model.TileCoordinateOf(int(worldX)>>8, int(worldY)>>8)
		tileX, tileY := coord.XY()

		if tileX >= 0 && tileX < 64 && tileY >= 0 && tileY < 64 {
			mode.mapDisplay.SetHighlightedTile(coord)
		} else {
			mode.mapDisplay.ClearHighlightedTile()
		}
		consumed = true
	}

	return
}

func (mode *LevelMapMode) onMouseButtonClicked(area *ui.Area, event events.Event) (consumed bool) {
	mouseEvent := event.(*events.MouseButtonEvent)

	if mouseEvent.AffectedButtons() == env.MousePrimary {
		worldX, worldY := mode.mapDisplay.WorldCoordinatesForPixel(mouseEvent.Position())
		coord := model.TileCoordinateOf(int(worldX)>>8, int(worldY)>>8)
		tileX, tileY := coord.XY()

		if tileX >= 0 && tileX < 64 && tileY >= 0 && tileY < 64 {
			if keys.Modifier(mouseEvent.Modifier()) == keys.ModControl {
				mode.toggleSelectedTile(coord)
			} else {
				mode.setSelectedTiles([]model.TileCoordinate{coord})
			}
			consumed = true
		}
	}

	return
}

func (mode *LevelMapMode) setSelectedTiles(tiles []model.TileCoordinate) {
	mode.selectedTiles = tiles
	mode.onSelectedTilesChanged()
}

func (mode *LevelMapMode) toggleSelectedTile(coord model.TileCoordinate) {
	newList := []model.TileCoordinate{}
	wasSelected := false

	for _, other := range mode.selectedTiles {
		if other != coord {
			newList = append(newList, other)
		} else {
			wasSelected = true
		}
	}
	if !wasSelected {
		newList = append(newList, coord)
	}

	mode.selectedTiles = newList
	mode.onSelectedTilesChanged()
}

func (mode *LevelMapMode) onSelectedTilesChanged() {
	mode.mapDisplay.SetSelectedTiles(mode.selectedTiles)
	tileMap := mode.context.ModelAdapter().ActiveLevel().TileMap()
	typeUnifier := util.NewValueUnifier(dataModel.TileType(""))

	for _, coord := range mode.selectedTiles {
		tile := tileMap.Tile(coord)
		properties := tile.Properties()
		if properties != nil {
			typeUnifier.Add(*properties.Type)
		}
	}
	mode.tileTypeBox.SetSelectedItem(mode.tileTypeItems[typeUnifier.Value().(dataModel.TileType)])
}

func (mode *LevelMapMode) onTilePropertyChangeRequested(item controls.ComboBoxItem) {
	propertyItem := item.(*tilePropertyItem)
	properties := &dataModel.TileProperties{}

	propertyItem.setter(properties, propertyItem.value)
	mode.context.ModelAdapter().ActiveLevel().RequestTilePropertyChange(mode.selectedTiles, properties)
}

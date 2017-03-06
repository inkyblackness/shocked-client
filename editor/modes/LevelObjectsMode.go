package modes

import (
	"sort"

	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/editor/display"
	"github.com/inkyblackness/shocked-client/editor/model"
	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/env/keys"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/ui"
	"github.com/inkyblackness/shocked-client/ui/events"
)

// LevelObjectsMode is a mode for level objects.
type LevelObjectsMode struct {
	context        Context
	levelAdapter   *model.LevelAdapter
	objectsAdapter *model.ObjectsAdapter

	displayFilter    func(*model.LevelObject) bool
	displayedObjects []*model.LevelObject

	mapDisplay *display.MapDisplay

	area       *ui.Area
	panel      *ui.Area
	panelRight ui.Anchor

	tileTypeLabel *controls.Label
	tileTypeBox   *controls.ComboBox

	closestObjects              []*model.LevelObject
	closestObjectHighlightIndex int
	selectedObjects             []*model.LevelObject
}

// NewLevelObjectsMode returns a new instance.
func NewLevelObjectsMode(context Context, parent *ui.Area, mapDisplay *display.MapDisplay) *LevelObjectsMode {
	mode := &LevelObjectsMode{
		context:        context,
		levelAdapter:   context.ModelAdapter().ActiveLevel(),
		objectsAdapter: context.ModelAdapter().ObjectsAdapter(),
		displayFilter:  func(*model.LevelObject) bool { return true },

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
		builder.OnEvent(events.MouseScrollEventType, mode.onMouseScrolled)
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
				graphics.RGBA(0.7, 0.0, 0.7, 0.1))
		})
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
		builder.OnEvent(events.MouseMoveEventType, func(area *ui.Area, event events.Event) (consumed bool) {
			moveEvent := event.(*events.MouseMoveEvent)
			if area.HasFocus() {
				newX, _ := moveEvent.Position()
				mode.panelRight.RequestValue(mode.panelRight.Value() + (newX - lastGrabX))
				lastGrabX = newX
				consumed = true
			}
			return
		})

		mode.panel = builder.Build()
	}

	mode.levelAdapter.OnLevelObjectsChanged(mode.onLevelObjectsChanged)

	return mode
}

// SetActive implements the Mode interface.
func (mode *LevelObjectsMode) SetActive(active bool) {
	if active {
		mode.updateDisplayedObjects()
		mode.onSelectedObjectsChanged()
	} else {
		mode.mapDisplay.SetDisplayedObjects(nil)
		mode.mapDisplay.SetHighlightedObject(nil)
		mode.mapDisplay.SetSelectedObjects(nil)
	}
	mode.area.SetVisible(active)
	mode.mapDisplay.SetVisible(active)
}

func (mode *LevelObjectsMode) onLevelObjectsChanged() {
	mode.updateDisplayedObjects()
}

func (mode *LevelObjectsMode) updateDisplayedObjects() {
	mode.displayedObjects = mode.levelAdapter.LevelObjects(mode.displayFilter)
	mode.mapDisplay.SetDisplayedObjects(mode.displayedObjects)

	mode.closestObjects = nil
	mode.closestObjectHighlightIndex = 0
	mode.updateClosestObjectHighlight()
}

func (mode *LevelObjectsMode) onMouseMoved(area *ui.Area, event events.Event) (consumed bool) {
	mouseEvent := event.(*events.MouseMoveEvent)

	if mouseEvent.Buttons() == 0 {
		worldX, worldY := mode.mapDisplay.WorldCoordinatesForPixel(mouseEvent.Position())
		mode.updateClosestDisplayedObjects(worldX, worldY)
		consumed = true
	}

	return
}

func (mode *LevelObjectsMode) onMouseScrolled(area *ui.Area, event events.Event) (consumed bool) {
	mouseEvent := event.(*events.MouseScrollEvent)

	if (mouseEvent.Buttons() == 0) && (keys.Modifier(mouseEvent.Modifier()) == keys.ModControl) {
		available := len(mode.closestObjects)

		if available > 1 {
			_, dy := mouseEvent.Deltas()
			delta := 1

			if dy < 0 {
				delta = -1
			}
			mode.closestObjectHighlightIndex = (available + mode.closestObjectHighlightIndex + delta) % available
			mode.updateClosestObjectHighlight()
		}
		consumed = true
	}

	return
}

func (mode *LevelObjectsMode) onMouseButtonClicked(area *ui.Area, event events.Event) (consumed bool) {
	mouseEvent := event.(*events.MouseButtonEvent)

	if mouseEvent.AffectedButtons() == env.MousePrimary {
		if len(mode.closestObjects) > 0 {
			object := mode.closestObjects[mode.closestObjectHighlightIndex]
			if keys.Modifier(mouseEvent.Modifier()) == keys.ModControl {
				mode.toggleSelectedObject(object)
			} else {
				mode.setSelectedObjects([]*model.LevelObject{object})
			}
		}
		consumed = true
	}

	return
}

func (mode *LevelObjectsMode) updateClosestDisplayedObjects(worldX, worldY float32) {
	type resultEntry struct {
		distance float32
		object   *model.LevelObject
	}
	entries := []*resultEntry{}
	refPoint := mgl.Vec2{worldX, worldY}
	limit := float32(48.0)

	for _, object := range mode.displayedObjects {
		otherX, otherY := object.Center()
		otherPoint := mgl.Vec2{otherX, otherY}
		delta := refPoint.Sub(otherPoint)
		len := delta.Len()

		if len <= limit {
			entries = append(entries, &resultEntry{len, object})
		}
	}
	sort.Slice(entries, func(a int, b int) bool { return entries[a].distance < entries[b].distance })
	mode.closestObjects = make([]*model.LevelObject, len(entries))
	for index, entry := range entries {
		mode.closestObjects[index] = entry.object
	}
	mode.closestObjectHighlightIndex = 0
	mode.updateClosestObjectHighlight()
}

func (mode *LevelObjectsMode) updateClosestObjectHighlight() {
	if len(mode.closestObjects) > 0 {
		mode.mapDisplay.SetHighlightedObject(mode.closestObjects[mode.closestObjectHighlightIndex])
	} else {
		mode.mapDisplay.SetHighlightedObject(nil)
	}
}

func (mode *LevelObjectsMode) setSelectedObjects(objects []*model.LevelObject) {
	mode.selectedObjects = objects
	mode.onSelectedObjectsChanged()
}

func (mode *LevelObjectsMode) toggleSelectedObject(object *model.LevelObject) {
	newList := []*model.LevelObject{}
	wasSelected := false

	for _, other := range mode.selectedObjects {
		if other.Index() != object.Index() {
			newList = append(newList, other)
		} else {
			wasSelected = true
		}
	}
	if !wasSelected {
		newList = append(newList, object)
	}
	mode.setSelectedObjects(newList)
}

func (mode *LevelObjectsMode) onSelectedObjectsChanged() {
	mode.mapDisplay.SetSelectedObjects(mode.selectedObjects)
}

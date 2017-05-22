package modes

import (
	"fmt"
	"math"
	"sort"
	"strings"

	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/res"
	"github.com/inkyblackness/res/data/interpreters"
	"github.com/inkyblackness/res/data/levelobj"

	dataModel "github.com/inkyblackness/shocked-model"

	"github.com/inkyblackness/shocked-client/editor/display"
	"github.com/inkyblackness/shocked-client/editor/model"
	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/env/keys"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/ui"
	"github.com/inkyblackness/shocked-client/ui/events"
	"github.com/inkyblackness/shocked-client/util"
)

var classNames = []string{
	"Weapons 0",
	"AmmoClips 1",
	"Projectiles 2",
	"Explosives 3",
	"Patches 4",
	"Hardware 5",
	"Software 6",
	"Scenery 7",
	"Items 8",
	"Panels 9",
	"Barriers 10",
	"Animations 11",
	"Markers 12",
	"Containers 13",
	"Critters 14"}

var maxObjectsPerClass = []int{16, 32, 32, 32, 32, 8, 16, 176, 128, 64, 64, 32, 160, 64, 64}

type newObjectClassItem struct {
	class int
}

func (item *newObjectClassItem) String() string {
	return classNames[item.class]
}

type objectTypeItem struct {
	id          model.ObjectID
	displayName string
}

func (item *objectTypeItem) String() string {
	return item.displayName + " (" + item.id.String() + ")"
}

type tabItem struct {
	area        *ui.Area
	displayName string
}

func (item *tabItem) String() string {
	return item.displayName
}

type enumItem struct {
	value       uint32
	displayName string
}

func (item *enumItem) String() string {
	return item.displayName
}

type disposableControl interface {
	Dispose()
}

type levelObjectProperty struct {
	title *controls.Label
	value disposableControl
}

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

	closestObjects              []*model.LevelObject
	closestObjectHighlightIndex int
	selectedObjects             []*model.LevelObject

	newObjectID model.ObjectID

	newObjectClassLabel        *controls.Label
	newObjectClassBox          *controls.ComboBox
	objectClassUsageTitleLabel *controls.Label
	objectClassUsageInfoLabel  *controls.Label
	newObjectTypeLabel         *controls.Label
	newObjectTypeBox           *controls.ComboBox

	highlightedObjectInfoTitle *controls.Label
	highlightedObjectInfoValue *controls.Label

	selectedObjectsTitleLabel   *controls.Label
	selectedObjectsDeleteLabel  *controls.Label
	selectedObjectsDeleteButton *controls.TextButton
	selectedObjectsIDTitleLabel *controls.Label
	selectedObjectsIDInfoLabel  *controls.Label
	selectedObjectsTypeLabel    *controls.Label
	selectedObjectsTypeBox      *controls.ComboBox

	selectedObjectsPropertiesTitle *controls.Label
	selectedObjectsPropertiesBox   *controls.ComboBox

	selectedObjectsBasePropertiesItem *tabItem
	selectedObjectsBasePropertiesArea *ui.Area
	selectedObjectsZTitle             *controls.Label
	selectedObjectsZValue             *controls.Slider
	selectedObjectsTileXTitle         *controls.Label
	selectedObjectsTileXValue         *controls.Slider
	selectedObjectsFineXTitle         *controls.Label
	selectedObjectsFineXValue         *controls.Slider
	selectedObjectsTileYTitle         *controls.Label
	selectedObjectsTileYValue         *controls.Slider
	selectedObjectsFineYTitle         *controls.Label
	selectedObjectsFineYValue         *controls.Slider
	selectedObjectsRotationXTitle     *controls.Label
	selectedObjectsRotationXValue     *controls.Slider
	selectedObjectsRotationYTitle     *controls.Label
	selectedObjectsRotationYValue     *controls.Slider
	selectedObjectsRotationZTitle     *controls.Label
	selectedObjectsRotationZValue     *controls.Slider
	selectedObjectsHitpointsTitle     *controls.Label
	selectedObjectsHitpointsValue     *controls.Slider

	selectedObjectsClassPropertiesItem    *tabItem
	selectedObjectsPropertiesMainArea     *ui.Area
	selectedObjectsPropertiesHeaderArea   *ui.Area
	selectedObjectsPropertiesArea         *ui.Area
	selectedObjectsPropertiesPanelBuilder *controlPanelBuilder
	selectedObjectsPropertiesBottom       ui.Anchor
	selectedObjectsProperties             []*levelObjectProperty
}

func intAsPointer(value int) (ptr *int) {
	ptr = new(int)
	*ptr = value
	return
}

// NewLevelObjectsMode returns a new instance.
func NewLevelObjectsMode(context Context, parent *ui.Area, mapDisplay *display.MapDisplay) *LevelObjectsMode {
	mode := &LevelObjectsMode{
		context:        context,
		levelAdapter:   context.ModelAdapter().ActiveLevel(),
		objectsAdapter: context.ModelAdapter().ObjectsAdapter(),
		displayFilter:  func(*model.LevelObject) bool { return true },

		newObjectID: model.MakeObjectID(0, 0, 0),

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
		mode.panelRight = ui.NewLimitedAnchor(minRight, maxRight, ui.NewOffsetAnchor(mode.area.Left(), 400))
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
		grabbing := false
		builder.OnEvent(events.MouseButtonDownEventType, func(area *ui.Area, event events.Event) bool {
			buttonEvent := event.(*events.MouseButtonEvent)
			if buttonEvent.Buttons() == env.MousePrimary {
				area.RequestFocus()
				lastGrabX, _ = buttonEvent.Position()
				grabbing = true
			}
			return true
		})
		builder.OnEvent(events.MouseButtonUpEventType, func(area *ui.Area, event events.Event) bool {
			buttonEvent := event.(*events.MouseButtonEvent)
			if buttonEvent.AffectedButtons() == env.MousePrimary {
				area.ReleaseFocus()
				grabbing = false
			}
			return true
		})
		builder.OnEvent(events.MouseMoveEventType, func(area *ui.Area, event events.Event) bool {
			moveEvent := event.(*events.MouseMoveEvent)
			if grabbing {
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

		mode.newObjectClassLabel, mode.newObjectClassBox = panelBuilder.addComboProperty("New Object Class", mode.onNewObjectClassChanged)
		mode.objectClassUsageTitleLabel, mode.objectClassUsageInfoLabel = panelBuilder.addInfo("Class Usage")
		mode.newObjectTypeLabel, mode.newObjectTypeBox = panelBuilder.addComboProperty("New Object Type", mode.onNewObjectTypeChanged)

		mode.highlightedObjectInfoTitle, mode.highlightedObjectInfoValue = panelBuilder.addInfo("Highlighted Object")

		mode.selectedObjectsTitleLabel = panelBuilder.addTitle("Selected Object(s)")
		mode.selectedObjectsDeleteLabel, mode.selectedObjectsDeleteButton = panelBuilder.addTextButton("Delete Selected", "Delete", mode.deleteSelectedObjects)
		mode.selectedObjectsIDTitleLabel, mode.selectedObjectsIDInfoLabel = panelBuilder.addInfo("Object ID")
		mode.selectedObjectsTypeLabel, mode.selectedObjectsTypeBox = panelBuilder.addComboProperty("Type", func(item controls.ComboBoxItem) {
			typeItem := item.(*objectTypeItem)
			mode.updateSelectedObjectsBaseProperties(func(properties *dataModel.LevelObjectProperties) {
				properties.Subclass = intAsPointer(typeItem.id.Subclass())
				properties.Type = intAsPointer(typeItem.id.Type())
			})
		})

		mode.selectedObjectsPropertiesTitle, mode.selectedObjectsPropertiesBox = panelBuilder.addComboProperty("Show Properties", mode.onSelectedPropertiesDisplayChanged)

		var basePropertiesPanelBuilder *controlPanelBuilder
		mode.selectedObjectsBasePropertiesArea, basePropertiesPanelBuilder = panelBuilder.addSection(false)
		mode.selectedObjectsZTitle, mode.selectedObjectsZValue = basePropertiesPanelBuilder.addSliderProperty("Z", func(newValue int64) {
			mode.updateSelectedObjectsBaseProperties(func(properties *dataModel.LevelObjectProperties) {
				properties.Z = intAsPointer(int(newValue))
			})
		})
		mode.selectedObjectsZValue.SetRange(0, 255)

		mode.selectedObjectsTileXTitle, mode.selectedObjectsTileXValue = basePropertiesPanelBuilder.addSliderProperty("TileX", func(newValue int64) {
			mode.updateSelectedObjectsBaseProperties(func(properties *dataModel.LevelObjectProperties) {
				properties.TileX = intAsPointer(int(newValue))
			})
		})
		mode.selectedObjectsTileXValue.SetRange(0, 63)
		mode.selectedObjectsFineXTitle, mode.selectedObjectsFineXValue = basePropertiesPanelBuilder.addSliderProperty("FineX", func(newValue int64) {
			mode.updateSelectedObjectsBaseProperties(func(properties *dataModel.LevelObjectProperties) {
				properties.FineX = intAsPointer(int(newValue))
			})
		})
		mode.selectedObjectsFineXValue.SetRange(0, 255)

		mode.selectedObjectsTileYTitle, mode.selectedObjectsTileYValue = basePropertiesPanelBuilder.addSliderProperty("TileY", func(newValue int64) {
			mode.updateSelectedObjectsBaseProperties(func(properties *dataModel.LevelObjectProperties) {
				properties.TileY = intAsPointer(int(newValue))
			})
		})
		mode.selectedObjectsTileYValue.SetRange(0, 63)
		mode.selectedObjectsFineYTitle, mode.selectedObjectsFineYValue = basePropertiesPanelBuilder.addSliderProperty("FineY", func(newValue int64) {
			mode.updateSelectedObjectsBaseProperties(func(properties *dataModel.LevelObjectProperties) {
				properties.FineY = intAsPointer(int(newValue))
			})
		})
		mode.selectedObjectsFineYValue.SetRange(0, 255)

		mode.selectedObjectsRotationXTitle, mode.selectedObjectsRotationXValue = basePropertiesPanelBuilder.addSliderProperty("RotationX", func(newValue int64) {
			mode.updateSelectedObjectsBaseProperties(func(properties *dataModel.LevelObjectProperties) {
				properties.RotationX = intAsPointer(int(newValue))
			})
		})
		mode.selectedObjectsRotationXValue.SetRange(0, 255)
		mode.selectedObjectsRotationYTitle, mode.selectedObjectsRotationYValue = basePropertiesPanelBuilder.addSliderProperty("RotationY", func(newValue int64) {
			mode.updateSelectedObjectsBaseProperties(func(properties *dataModel.LevelObjectProperties) {
				properties.RotationY = intAsPointer(int(newValue))
			})
		})
		mode.selectedObjectsRotationYValue.SetRange(0, 255)
		mode.selectedObjectsRotationZTitle, mode.selectedObjectsRotationZValue = basePropertiesPanelBuilder.addSliderProperty("RotationZ", func(newValue int64) {
			mode.updateSelectedObjectsBaseProperties(func(properties *dataModel.LevelObjectProperties) {
				properties.RotationZ = intAsPointer(int(newValue))
			})
		})
		mode.selectedObjectsRotationZValue.SetRange(0, 255)

		mode.selectedObjectsHitpointsTitle, mode.selectedObjectsHitpointsValue = basePropertiesPanelBuilder.addSliderProperty("Hitpoints", func(newValue int64) {
			mode.updateSelectedObjectsBaseProperties(func(properties *dataModel.LevelObjectProperties) {
				properties.Hitpoints = intAsPointer(int(newValue))
			})
		})
		mode.selectedObjectsHitpointsValue.SetRange(0, 10000)

		classPropertiesBottomResolver := func() ui.Anchor { return mode.selectedObjectsPropertiesBottom }
		var mainClassPanelBuilder *controlPanelBuilder
		mode.selectedObjectsPropertiesMainArea, mainClassPanelBuilder =
			panelBuilder.addDynamicSection(true, classPropertiesBottomResolver)
		mode.selectedObjectsPropertiesHeaderArea, _ = mainClassPanelBuilder.addSection(true)

		mode.selectedObjectsPropertiesArea, mode.selectedObjectsPropertiesPanelBuilder =
			mainClassPanelBuilder.addDynamicSection(true, classPropertiesBottomResolver)
		mode.selectedObjectsPropertiesBottom = mode.selectedObjectsPropertiesHeaderArea.Bottom()

		mode.selectedObjectsBasePropertiesItem = &tabItem{mode.selectedObjectsBasePropertiesArea, "Base Properties"}
		mode.selectedObjectsClassPropertiesItem = &tabItem{mode.selectedObjectsPropertiesMainArea, "Class Properties"}
		propertiesTabItems := []controls.ComboBoxItem{mode.selectedObjectsBasePropertiesItem, mode.selectedObjectsClassPropertiesItem}
		mode.selectedObjectsPropertiesBox.SetItems(propertiesTabItems)
		mode.selectedObjectsPropertiesBox.SetSelectedItem(mode.selectedObjectsClassPropertiesItem)
	}

	mode.levelAdapter.OnLevelObjectsChanged(mode.onLevelObjectsChanged)
	mode.context.ModelAdapter().ObjectsAdapter().OnObjectsChanged(mode.onGameObjectsChanged)

	return mode
}

// SetActive implements the Mode interface.
func (mode *LevelObjectsMode) SetActive(active bool) {
	if active {
		mode.updateDisplayedObjects()
		mode.mapDisplay.SetSelectedObjects(mode.selectedObjects)
	} else {
		mode.mapDisplay.SetDisplayedObjects(nil)
		mode.mapDisplay.SetHighlightedObject(nil)
		mode.mapDisplay.SetSelectedObjects(nil)
	}
	mode.area.SetVisible(active)
	mode.mapDisplay.SetVisible(active)
}

func (mode *LevelObjectsMode) onLevelObjectsChanged() {
	mode.updateNewObjectClassQuota()
	if mode.area.IsVisible() {
		mode.updateDisplayedObjects()
	}
}

func (mode *LevelObjectsMode) onSelectedPropertiesDisplayChanged(item controls.ComboBoxItem) {
	tabItem := item.(*tabItem)

	mode.selectedObjectsBasePropertiesItem.area.SetVisible(false)
	mode.selectedObjectsClassPropertiesItem.area.SetVisible(false)
	tabItem.area.SetVisible(true)
}

func (mode *LevelObjectsMode) updateDisplayedObjects() {
	mode.displayedObjects = mode.levelAdapter.LevelObjects(mode.displayFilter)
	mode.mapDisplay.SetDisplayedObjects(mode.displayedObjects)

	mode.closestObjects = nil
	mode.closestObjectHighlightIndex = 0
	mode.updateClosestObjectHighlight()

	mode.updateSelectedFromDisplayedObjects()
}

func (mode *LevelObjectsMode) updateSelectedFromDisplayedObjects() {
	displayedIndices := make(map[int]bool)
	for _, displayedObject := range mode.displayedObjects {
		displayedIndices[displayedObject.Index()] = true
	}

	selectedObjects := make([]*model.LevelObject, 0, len(mode.selectedObjects))
	for _, selectedObject := range mode.selectedObjects {
		if displayedIndices[selectedObject.Index()] {
			selectedObjects = append(selectedObjects, selectedObject)
		}
	}
	mode.setSelectedObjects(selectedObjects)
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
	} else if mouseEvent.AffectedButtons() == env.MouseSecondary {
		worldX, worldY := mode.mapDisplay.WorldCoordinatesForPixel(mouseEvent.Position())
		mode.createNewObject(worldX, worldY)
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
		object := mode.closestObjects[mode.closestObjectHighlightIndex]
		mode.highlightedObjectInfoValue.SetText(fmt.Sprintf("%v: %v (%v)", object.Index(), object.ID(), mode.objectDisplayName(object.ID())))
		mode.mapDisplay.SetHighlightedObject(object)
	} else {
		mode.highlightedObjectInfoValue.SetText("")
		mode.mapDisplay.SetHighlightedObject(nil)
	}
}

func (mode *LevelObjectsMode) objectDisplayName(id model.ObjectID) string {
	displayName := "unknown"

	if gameObject := mode.context.ModelAdapter().ObjectsAdapter().Object(id); gameObject != nil {
		displayName = gameObject.DisplayName()
	}

	return displayName
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

	classUnifier := util.NewValueUnifier(-1)
	subclassUnifier := util.NewValueUnifier(-1)
	typeUnifier := util.NewValueUnifier(-1)
	tileXUnifier := util.NewValueUnifier(-1)
	fineXUnifier := util.NewValueUnifier(-1)
	tileYUnifier := util.NewValueUnifier(-1)
	fineYUnifier := util.NewValueUnifier(-1)
	zUnifier := util.NewValueUnifier(-1)
	rotationXUnifier := util.NewValueUnifier(-1)
	rotationYUnifier := util.NewValueUnifier(-1)
	rotationZUnifier := util.NewValueUnifier(-1)
	hitpointsUnifier := util.NewValueUnifier(-1)

	for _, object := range mode.selectedObjects {
		classUnifier.Add(object.ID().Class())
		subclassUnifier.Add(object.ID().Subclass())
		typeUnifier.Add(object.ID().Type())
		zUnifier.Add(object.Z())
		tileXUnifier.Add(object.TileX())
		fineXUnifier.Add(object.FineX())
		tileYUnifier.Add(object.TileY())
		fineYUnifier.Add(object.FineY())
		rotationXUnifier.Add(object.RotationX())
		rotationYUnifier.Add(object.RotationY())
		rotationZUnifier.Add(object.RotationZ())
		hitpointsUnifier.Add(object.Hitpoints())
	}
	unifiedClass := classUnifier.Value().(int)
	unifiedSubclass := subclassUnifier.Value().(int)
	unifiedType := typeUnifier.Value().(int)
	var unifiedIDString string
	if unifiedClass != -1 {
		unifiedIDString = classNames[unifiedClass]
		typeItems := mode.objectItemsForClass(unifiedClass)
		mode.selectedObjectsTypeBox.SetItems(typeItems)
	} else {
		unifiedIDString = "**"
		unifiedSubclass = -1
		mode.selectedObjectsTypeBox.SetItems(nil)
	}
	unifiedIDString += "/"
	if unifiedSubclass != -1 {
		unifiedIDString += fmt.Sprintf("%v", unifiedSubclass)
	} else {
		unifiedIDString += "*"
		unifiedType = -1
	}
	unifiedIDString += "/"
	if unifiedType != -1 {
		unifiedIDString += fmt.Sprintf("%v", unifiedType)
	} else {
		unifiedIDString += "*"
	}
	if len(mode.selectedObjects) > 0 {
		mode.selectedObjectsIDInfoLabel.SetText(unifiedIDString)
	} else {
		mode.selectedObjectsIDInfoLabel.SetText("")
	}
	if (unifiedClass != -1) && (unifiedSubclass != -1) && (unifiedType != -1) {
		objectID := model.MakeObjectID(unifiedClass, unifiedSubclass, unifiedType)
		item := &objectTypeItem{objectID, mode.objectDisplayName(objectID)}
		mode.selectedObjectsTypeBox.SetSelectedItem(item)
	} else {
		mode.selectedObjectsTypeBox.SetSelectedItem(nil)
	}

	setSliderValue := func(slider *controls.Slider, unifier *util.ValueUnifier) {
		value := unifier.Value().(int)
		if value >= 0 {
			slider.SetValue(int64(value))
		} else {
			slider.SetValueUndefined()
		}
	}

	setSliderValue(mode.selectedObjectsTileXValue, tileXUnifier)
	setSliderValue(mode.selectedObjectsFineXValue, fineXUnifier)
	setSliderValue(mode.selectedObjectsTileYValue, tileYUnifier)
	setSliderValue(mode.selectedObjectsFineYValue, fineYUnifier)
	setSliderValue(mode.selectedObjectsZValue, zUnifier)
	setSliderValue(mode.selectedObjectsRotationXValue, rotationXUnifier)
	setSliderValue(mode.selectedObjectsRotationYValue, rotationYUnifier)
	setSliderValue(mode.selectedObjectsRotationZValue, rotationZUnifier)
	setSliderValue(mode.selectedObjectsHitpointsValue, hitpointsUnifier)

	mode.recreateLevelObjectProperties()
}

func (mode *LevelObjectsMode) recreateLevelObjectProperties() {
	for _, oldProperty := range mode.selectedObjectsProperties {
		oldProperty.title.Dispose()
		oldProperty.value.Dispose()
	}
	mode.selectedObjectsPropertiesPanelBuilder.reset()
	mode.selectedObjectsPropertiesBottom = mode.selectedObjectsPropertiesHeaderArea.Bottom()

	var newProperties = []*levelObjectProperty{}
	if len(mode.selectedObjects) > 0 {
		propertyUnifier := make(map[string]*util.ValueUnifier)
		propertyDescribers := make(map[string]func(*interpreters.Simplifier))
		propertyOrder := []string{}
		describer := func(interpreter *interpreters.Instance, key string) func(simpl *interpreters.Simplifier) {
			return func(simpl *interpreters.Simplifier) { interpreter.Describe(key, simpl) }
		}

		var unifyInterpreter func(string, *interpreters.Instance, bool, map[string]bool)
		unifyInterpreter = func(path string, interpreter *interpreters.Instance, first bool, thisKeys map[string]bool) {
			for _, key := range interpreter.Keys() {
				fullPath := path + key
				thisKeys[fullPath] = true
				if unifier, existing := propertyUnifier[fullPath]; existing || first {
					if !existing {
						unifier = util.NewValueUnifier(int64(math.MinInt64))
						propertyUnifier[fullPath] = unifier
						propertyDescribers[fullPath] = describer(interpreter, key)
						propertyOrder = append(propertyOrder, fullPath)
					}
					unifier.Add(int64(interpreter.Get(key)))
				}
			}
			for _, key := range interpreter.ActiveRefinements() {
				unifyInterpreter(path+key+".", interpreter.Refined(key), first, thisKeys)
			}
		}

		interpreterFactory := mode.interpreterFactory()

		for index, object := range mode.selectedObjects {
			objID := object.ID()
			resID := res.MakeObjectID(res.ObjectClass(objID.Class()), res.ObjectSubclass(objID.Subclass()), res.ObjectType(objID.Type()))
			interpreter := interpreterFactory(resID, object.ClassData())
			thisKeys := make(map[string]bool)
			unifyInterpreter("", interpreter, index == 0, thisKeys)
			{
				toRemove := []string{}
				for previousKey := range propertyUnifier {
					if !thisKeys[previousKey] {
						toRemove = append(toRemove, previousKey)
					}
				}
				for _, key := range toRemove {
					delete(propertyUnifier, key)
				}
			}
		}

		for _, key := range propertyOrder {
			if unifier, existing := propertyUnifier[key]; existing {
				properties := mode.createPropertyControls(key, unifier, propertyDescribers[key])
				newProperties = append(newProperties, properties...)
				mode.selectedObjectsPropertiesBottom = mode.selectedObjectsPropertiesPanelBuilder.bottom()
			}
		}
	}
	mode.selectedObjectsProperties = newProperties
}

func (mode *LevelObjectsMode) interpreterFactory() func(resID res.ObjectID, classData []byte) *interpreters.Instance {
	factory := levelobj.ForRealWorld
	if mode.levelAdapter.IsCyberspace() {
		factory = levelobj.ForCyberspace
	}
	return factory
}

func (mode *LevelObjectsMode) createPropertyControls(key string,
	unifier *util.ValueUnifier, describer func(*interpreters.Simplifier)) (properties []*levelObjectProperty) {
	unifiedValue := unifier.Value().(int64)

	simplifier := interpreters.NewSimplifier(func(minValue, maxValue int64) {
		title, slider := mode.selectedObjectsPropertiesPanelBuilder.addSliderProperty(key, func(newValue int64) {
			mode.updateSelectedObjectsClassProperties(key, uint32(newValue))
		})
		slider.SetRange(minValue, maxValue)
		if unifiedValue != math.MinInt64 {
			slider.SetValue(unifiedValue)
		}
		properties = append(properties, &levelObjectProperty{title, slider})
	})

	simplifier.SetEnumValueHandler(func(values map[uint32]string) {
		title, box := mode.selectedObjectsPropertiesPanelBuilder.addComboProperty(key, func(item controls.ComboBoxItem) {
			enumItem := item.(*enumItem)
			mode.updateSelectedObjectsClassProperties(key, enumItem.value)
		})
		valueKeys := make([]uint32, 0, len(values))
		for valueKey := range values {
			valueKeys = append(valueKeys, valueKey)
		}
		sort.Slice(valueKeys, func(indexA, indexB int) bool { return valueKeys[indexA] < valueKeys[indexB] })
		items := make([]controls.ComboBoxItem, len(valueKeys))
		var selectedItem controls.ComboBoxItem
		for index, valueKey := range valueKeys {
			items[index] = &enumItem{valueKey, values[valueKey]}
			if int64(valueKey) == unifiedValue {
				selectedItem = items[index]
			}
		}
		box.SetItems(items)
		box.SetSelectedItem(selectedItem)
		properties = append(properties, &levelObjectProperty{title, box})
	})

	simplifier.SetObjectIndexHandler(func() {
		title, slider := mode.selectedObjectsPropertiesPanelBuilder.addSliderProperty(key, func(newValue int64) {
			mode.updateSelectedObjectsClassProperties(key, uint32(newValue))
		})
		slider.SetRange(0, 871)
		if unifiedValue != math.MinInt64 {
			slider.SetValue(unifiedValue)
		}
		properties = append(properties, &levelObjectProperty{title, slider})
	})

	addVariableKey := func() {
		typeTitle, typeBox := mode.selectedObjectsPropertiesPanelBuilder.addComboProperty(key+"-Type", func(item controls.ComboBoxItem) {
			enumItem := item.(*enumItem)
			mode.updateSelectedObjectsClassPropertiesFiltered(key, enumItem.value, 0, 0x1000)
		})
		items := make([]controls.ComboBoxItem, 2)
		items[0] = &enumItem{0, "Boolean"}
		items[1] = &enumItem{0x1000, "Integer"}
		var selectedItem controls.ComboBoxItem
		if unifiedValue != math.MinInt64 {
			if (unifiedValue & 0x1000) != 0 {
				selectedItem = items[1]
			} else {
				selectedItem = items[0]
			}
		}
		typeBox.SetItems(items)
		typeBox.SetSelectedItem(selectedItem)

		properties = append(properties, &levelObjectProperty{typeTitle, typeBox})
		indexTitle, indexSlider := mode.selectedObjectsPropertiesPanelBuilder.addSliderProperty(key+"-Index", func(newValue int64) {
			mode.updateSelectedObjectsClassPropertiesFiltered(key, uint32(newValue), 0, 0x1FF)
		})
		indexSlider.SetRange(0, 0x1FF)
		if unifiedValue != math.MinInt64 {
			indexSlider.SetValue(unifiedValue & 0x1FF)
			if (unifiedValue & 0x1000) != 0 {
				indexSlider.SetRange(0, 0x3F)
			}
		}
		properties = append(properties, &levelObjectProperty{indexTitle, indexSlider})
	}

	simplifier.SetSpecialHandler("VariableKey", addVariableKey)
	simplifier.SetSpecialHandler("VariableCondition", func() {
		addVariableKey()

		comparisonTitle, comparisonBox := mode.selectedObjectsPropertiesPanelBuilder.addComboProperty(key+"-Check", func(item controls.ComboBoxItem) {
			enumItem := item.(*enumItem)
			mode.updateSelectedObjectsClassPropertiesFiltered(key, enumItem.value, 13, 0xE000)
		})
		var selectedItem controls.ComboBoxItem
		items := []controls.ComboBoxItem{
			&enumItem{0, "Var == Val"},
			&enumItem{1, "Var < Val"},
			&enumItem{2, "Var <= Val"},
			&enumItem{3, "Var > Val"},
			&enumItem{4, "Var >= Val"},
			&enumItem{5, "Var != Val"}}
		if unifiedValue != math.MinInt64 {
			selectedItem = items[unifiedValue>>13]
		}
		comparisonBox.SetItems(items)
		comparisonBox.SetSelectedItem(selectedItem)
		properties = append(properties, &levelObjectProperty{comparisonTitle, comparisonBox})
	})

	simplifier.SetSpecialHandler("BinaryCodedDecimal", func() {
		title, slider := mode.selectedObjectsPropertiesPanelBuilder.addSliderProperty(key, func(newValue int64) {
			mode.updateSelectedObjectsClassProperties(key, uint32(util.ToBinaryCodedDecimal(uint16(newValue))))
		})
		slider.SetRange(0, 999)
		if unifiedValue != math.MinInt64 {
			slider.SetValue(int64(util.FromBinaryCodedDecimal(uint16(unifiedValue))))
		}
		properties = append(properties, &levelObjectProperty{title, slider})
	})

	describer(simplifier)

	return properties
}

func (mode *LevelObjectsMode) selectedObjectIndices() []int {
	objectIndices := make([]int, len(mode.selectedObjects))
	for index, object := range mode.selectedObjects {
		objectIndices[index] = object.Index()
	}
	return objectIndices
}

func (mode *LevelObjectsMode) updateSelectedObjectsBaseProperties(modifier func(properties *dataModel.LevelObjectProperties)) {
	var properties dataModel.LevelObjectProperties
	modifier(&properties)
	mode.levelAdapter.RequestObjectPropertiesChange(mode.selectedObjectIndices(), &properties)
}

func (mode *LevelObjectsMode) updateSelectedObjectsClassProperties(key string, value uint32) {
	mode.updateSelectedObjectsClassPropertiesFiltered(key, value, 0, 0xFFFFFFFF)
	/*
		interpreterFactory := mode.interpreterFactory()

		for _, object := range mode.selectedObjects {
			objID := object.ID()
			resID := res.MakeObjectID(res.ObjectClass(objID.Class()), res.ObjectSubclass(objID.Subclass()), res.ObjectType(objID.Type()))
			var properties dataModel.LevelObjectProperties

			properties.ClassData = object.ClassData()
			interpreter := interpreterFactory(resID, properties.ClassData)
			subKeys := strings.Split(key, ".")
			valueIndex := len(subKeys) - 1
			for subIndex := 0; subIndex < valueIndex; subIndex++ {
				interpreter = interpreter.Refined(subKeys[subIndex])
			}
			interpreter.Set(subKeys[valueIndex], value)
			mode.levelAdapter.RequestObjectPropertiesChange([]int{object.Index()}, &properties)
		}
	*/
}

func (mode *LevelObjectsMode) updateSelectedObjectsClassPropertiesFiltered(key string, value uint32, offset uint32, mask uint32) {
	interpreterFactory := mode.interpreterFactory()

	for _, object := range mode.selectedObjects {
		objID := object.ID()
		resID := res.MakeObjectID(res.ObjectClass(objID.Class()), res.ObjectSubclass(objID.Subclass()), res.ObjectType(objID.Type()))
		var properties dataModel.LevelObjectProperties

		properties.ClassData = object.ClassData()
		interpreter := interpreterFactory(resID, properties.ClassData)
		subKeys := strings.Split(key, ".")
		valueIndex := len(subKeys) - 1
		for subIndex := 0; subIndex < valueIndex; subIndex++ {
			interpreter = interpreter.Refined(subKeys[subIndex])
		}
		interpreter.Set(subKeys[valueIndex], (value<<offset)|(interpreter.Get(subKeys[valueIndex]) & ^mask))
		mode.levelAdapter.RequestObjectPropertiesChange([]int{object.Index()}, &properties)
	}
}

func (mode *LevelObjectsMode) deleteSelectedObjects() {
	mode.levelAdapter.RequestRemoveObjects(mode.selectedObjectIndices())
}

func (mode *LevelObjectsMode) onGameObjectsChanged() {
	newClassItems := make([]controls.ComboBoxItem, len(classNames))

	for index := range classNames {
		newClassItems[index] = &newObjectClassItem{index}
	}
	mode.newObjectClassBox.SetItems(newClassItems)
	mode.newObjectClassBox.SetSelectedItem(newClassItems[mode.newObjectID.Class()])
	mode.updateNewObjectClass(mode.newObjectID.Class())
}

func (mode *LevelObjectsMode) onNewObjectClassChanged(item controls.ComboBoxItem) {
	classItem := item.(*newObjectClassItem)
	mode.updateNewObjectClass(classItem.class)
	mode.updateNewObjectClassQuota()
}

func (mode *LevelObjectsMode) updateNewObjectClass(objectClass int) {
	typeItems := mode.objectItemsForClass(objectClass)

	mode.newObjectTypeBox.SetItems(typeItems)
	if len(typeItems) > 0 {
		mode.newObjectTypeBox.SetSelectedItem(typeItems[0])
		mode.onNewObjectTypeChanged(typeItems[0])
	} else {
		mode.newObjectTypeBox.SetSelectedItem(nil)
		mode.newObjectID = model.MakeObjectID(objectClass, 0, 0)
	}
}

func (mode *LevelObjectsMode) updateNewObjectClassQuota() {
	maxCount := maxObjectsPerClass[mode.newObjectID.Class()] - 1 // zero entry can never be used, so it's one less
	currentClassObjects := mode.levelAdapter.LevelObjects(func(object *model.LevelObject) bool {
		return object.ID().Class() == mode.newObjectID.Class()
	})

	mode.objectClassUsageInfoLabel.SetText(fmt.Sprintf("%3d/%3d", len(currentClassObjects), maxCount))
}

func (mode *LevelObjectsMode) onNewObjectTypeChanged(item controls.ComboBoxItem) {
	typeItem := item.(*objectTypeItem)
	mode.newObjectID = typeItem.id
}

func (mode *LevelObjectsMode) createNewObject(worldX, worldY float32) {
	mode.levelAdapter.RequestNewObject(worldX, worldY, mode.newObjectID)
}

func (mode *LevelObjectsMode) objectItemsForClass(objectClass int) []controls.ComboBoxItem {
	availableGameObjects := mode.context.ModelAdapter().ObjectsAdapter().ObjectsOfClass(objectClass)
	typeItems := make([]controls.ComboBoxItem, len(availableGameObjects))

	for index, gameObject := range availableGameObjects {
		typeItems[index] = &objectTypeItem{gameObject.ID(), gameObject.DisplayName()}
	}

	return typeItems
}

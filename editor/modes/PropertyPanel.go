package modes

import (
	"math"
	"sort"

	"github.com/inkyblackness/res/data/interpreters"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/ui"
)

type disposableControl interface {
	Dispose()
}

type enumItem struct {
	value       uint32
	displayName string
}

func (item *enumItem) String() string {
	return item.displayName
}

type propertyEntry struct {
	title   *controls.Label
	control disposableControl
}

type propertyUpdateFunction func(currentValue, parameter uint32) uint32

func setUpdate() propertyUpdateFunction {
	return func(currentValue, parameter uint32) uint32 { return parameter }
}

func maskedUpdate(shift, mask uint32) propertyUpdateFunction {
	return func(currentValue, parameter uint32) uint32 {
		return (parameter << shift) | (currentValue & ^mask)
	}
}

type propertyChangeHandler func(key string, parameter uint32, update propertyUpdateFunction)

type propertyPanel struct {
	area          *ui.Area
	builder       *controlPanelBuilder
	changeHandler propertyChangeHandler
	entries       []*propertyEntry
}

func newPropertyPanel(parentBuilder *controlPanelBuilder, changeHandler propertyChangeHandler) *propertyPanel {
	panel := &propertyPanel{
		changeHandler: changeHandler}

	panel.area, panel.builder = parentBuilder.addDynamicSection(true, panel.Bottom)

	return panel
}

func (panel *propertyPanel) Bottom() ui.Anchor {
	return panel.builder.bottom()
}

func (panel *propertyPanel) Reset() {
	for _, entry := range panel.entries {
		entry.title.Dispose()
		entry.control.Dispose()
	}
	panel.entries = []*propertyEntry{}
	panel.builder.reset()
}

func (panel *propertyPanel) NewSimplifier(key string, unifiedValue int64) *interpreters.Simplifier {
	simplifier := interpreters.NewSimplifier(func(minValue, maxValue int64) {
		slider := panel.NewSlider(key, "", setUpdate())
		slider.SetRange(minValue, maxValue)
		if unifiedValue != math.MinInt64 {
			slider.SetValue(unifiedValue)
		}
	})

	simplifier.SetEnumValueHandler(func(values map[uint32]string) {
		box := panel.NewComboBox(key, "", setUpdate())
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
	})

	simplifier.SetBitfieldHandler(func(values map[uint32]string) {
		masks := make([]uint32, 0, len(values))

		for mask := range values {
			masks = append(masks, mask)
		}
		sort.Slice(masks, func(indexA, indexB int) bool { return masks[indexA] < masks[indexB] })
		for _, mask := range masks {
			maskName := values[mask]
			max := mask
			shift := uint32(0)

			for (max & 1) == 0 {
				shift++
				max >>= 1
			}
			slider := panel.NewSlider(key, maskName, maskedUpdate(shift, mask))
			slider.SetRange(0, int64(max))
			if unifiedValue != math.MinInt64 {
				slider.SetValue(int64((uint32(unifiedValue) & mask) >> shift))
			}
		}
	})

	return simplifier
}

func (panel *propertyPanel) fullName(key, nameSuffix string) (fullName string) {
	fullName = key
	if len(nameSuffix) > 0 {
		fullName += "-" + nameSuffix
	}
	return
}

func (panel *propertyPanel) NewSlider(key string, nameSuffix string, update propertyUpdateFunction) *controls.Slider {
	fullName := panel.fullName(key, nameSuffix)
	title, control := panel.builder.addSliderProperty(fullName, func(newValue int64) {
		panel.changeHandler(key, uint32(newValue), update)
	})

	panel.entries = append(panel.entries, &propertyEntry{title, control})

	return control
}

func (panel *propertyPanel) NewComboBox(key string, nameSuffix string, update propertyUpdateFunction) *controls.ComboBox {
	fullName := panel.fullName(key, nameSuffix)
	title, control := panel.builder.addComboProperty(fullName, func(item controls.ComboBoxItem) {
		enumItem := item.(*enumItem)
		panel.changeHandler(key, enumItem.value, update)
	})

	panel.entries = append(panel.entries, &propertyEntry{title, control})

	return control
}

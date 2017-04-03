package controls

import (
	"fmt"

	//mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/ui"
	"github.com/inkyblackness/shocked-client/ui/events"
)

// SliderChangeHandler is a callback for notifying the current value.
type SliderChangeHandler func(value int64)

// Slider is a control for selecting a numerical value with a slider.
type Slider struct {
	area         *ui.Area
	rectRenderer *graphics.RectangleRenderer

	valueLabel *Label

	sliderChangeHandler SliderChangeHandler

	valueMin int64
	valueMax int64

	valueUndefined bool
	value          int64
}

// Dispose releases all resources and removes the area from the tree.
func (slider *Slider) Dispose() {
	slider.valueLabel.Dispose()
	slider.area.Remove()
}

// SetRange sets the minimum and maximum of valid values.
func (slider *Slider) SetRange(min, max int64) {
	slider.valueMin, slider.valueMax = min, max
}

// SetValueUndefined clears the current value.
func (slider *Slider) SetValueUndefined() {
	slider.valueUndefined = true
	slider.value = 0
	slider.valueLabel.SetText("")
}

// SetValue updates the current value.
func (slider *Slider) SetValue(value int64) {
	slider.valueUndefined = false
	slider.value = value
	slider.valueLabel.SetText(fmt.Sprintf("%v", value))
}

func (slider *Slider) onRender(area *ui.Area) {
	slider.rectRenderer.Fill(area.Left().Value(), area.Top().Value(), area.Right().Value(), area.Bottom().Value(),
		graphics.RGBA(0.31, 0.56, 0.34, 0.8))
}

func (slider *Slider) onMouseButtonDown(area *ui.Area, event events.Event) bool {
	return true
}

func (slider *Slider) onMouseButtonUp(area *ui.Area, event events.Event) bool {
	return true
}

func (slider *Slider) onMouseMove(area *ui.Area, event events.Event) bool {
	return true
}

func (slider *Slider) onMouseScroll(area *ui.Area, event events.Event) bool {
	return true
}

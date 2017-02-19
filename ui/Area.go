package ui

import (
	"github.com/inkyblackness/shocked-client/ui/events"
)

// RenderFunction is called when an area wants to render its content.
type RenderFunction func(*Area, Renderer)

// EventHandler is called for events dispatched to the area.
type EventHandler func(*Area, events.Event) bool

// Area specifies one rectangular area within the user-interface stack.
type Area struct {
	parent   *Area
	children []*Area

	left   Anchor
	top    Anchor
	right  Anchor
	bottom Anchor

	onRender     RenderFunction
	eventHandler map[events.EventType]EventHandler
}

// Left returns the left anchor.
func (area *Area) Left() Anchor {
	return area.left
}

// Top returns the top anchor.
func (area *Area) Top() Anchor {
	return area.top
}

// Right returns the right anchor.
func (area *Area) Right() Anchor {
	return area.right
}

// Bottom returns the bottom anchor.
func (area *Area) Bottom() Anchor {
	return area.bottom
}

// Render first renders this area, then sequentially all children.
func (area *Area) Render(renderer Renderer) {
	area.onRender(area, renderer)
	for _, child := range area.children {
		child.Render(renderer)
	}
}

// HandleEvent tries to process the given event.
// It returns true if the area consumed the event.
func (area *Area) HandleEvent(event events.Event) bool {
	handler, existing := area.eventHandler[event.EventType()]
	result := false

	if existing {
		result = handler(area, event)
	}

	return result
}

// DispatchPositionalEvent tries to find an event handler in this areas
// UI tree at the position of the event. The event is tried depth-first,
// before trying to handle it within this area.
func (area *Area) DispatchPositionalEvent(event *events.MouseEvent) bool {
	x, y := event.Position()
	result := false

	for childIndex := len(area.children) - 1; !result && (childIndex >= 0); childIndex-- {
		child := area.children[childIndex]
		if (x >= child.Left().Value()) && (x < child.Right().Value()) &&
			(y >= child.Top().Value()) && (y < child.Bottom().Value()) {
			result = child.DispatchPositionalEvent(event)
		}
	}
	if !result {
		result = area.HandleEvent(event)
	}

	return result
}

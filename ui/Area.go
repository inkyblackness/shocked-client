package ui

// RenderFunction is called when an area wants to render its content.
type RenderFunction func(*Area, Renderer)

// Area specifies one rectangular area within the user-interface stack.
type Area struct {
	parent   *Area
	children []*Area

	left   Anchor
	top    Anchor
	right  Anchor
	bottom Anchor

	onRender RenderFunction
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
}
